package store

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

var MAGIC_NUMBER = []byte{42, 249, 180, 108, 82, 75, 222, 182}

var FILE_FORMAT_VERSION = []byte{1}

var MESSAGES_PER_FILE = uint64(10000)

type MessagePartition struct {
	basedir                 string
	name                    string
	appendFile              *os.File
	appendIndexFile         *os.File
	appendFirstId           uint64
	appendLastId            uint64
	appendFileWritePosition uint64
	maxMessageId            uint64
}

func NewMessagePartition(basedir string, storeName string) *MessagePartition {
	return &MessagePartition{
		basedir: basedir,
		name:    storeName,
	}
}

func (p *MessagePartition) Start() error {
	//fileList, err := p.scanFiles()
	return nil
}

// Returns the start messages ids for all available message files
// in a sorted list
func (p *MessagePartition) scanFiles() ([]uint64, error) {
	return nil, nil
}

func (p *MessagePartition) MaxMessageId() (uint64, error) {
	return p.maxMessageId, nil
}

func (p *MessagePartition) closeAppendFiles() error {
	if p.appendFile != nil {
		if err := p.appendFile.Close(); err != nil {
			if p.appendIndexFile != nil {
				defer p.appendIndexFile.Close()
			}
			return err
		}
	}

	if p.appendIndexFile != nil {
		return p.appendIndexFile.Close()
	}
	return nil
}

func (p *MessagePartition) createNextAppendFiles(msgId uint64) error {

	firstMessageIdForFile := p.firstMessageIdForFile(msgId)

	file, err := os.OpenFile(p.filenameByMessageId(firstMessageIdForFile), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// write file header on new files
	if stat, _ := file.Stat(); stat.Size() == 0 {
		p.appendFileWritePosition = uint64(stat.Size())

		_, err = file.Write(MAGIC_NUMBER)
		if err != nil {
			return err
		}

		_, err = file.Write(FILE_FORMAT_VERSION)
		if err != nil {
			return err
		}
	}

	index, errIndex := os.OpenFile(p.indexFilenameByMessageId(firstMessageIdForFile), os.O_RDWR|os.O_CREATE, 0666)
	if errIndex != nil {
		defer file.Close()
		defer os.Remove(file.Name())
		return err
	}

	p.appendFile = file
	p.appendIndexFile = index
	p.appendFirstId = firstMessageIdForFile
	p.appendLastId = firstMessageIdForFile + MESSAGES_PER_FILE - 1
	stat, err := file.Stat()
	p.appendFileWritePosition = uint64(stat.Size())
	return nil
}

func (p *MessagePartition) Close() error {
	return p.closeAppendFiles()
}

func (p *MessagePartition) Store(msgId uint64, msg []byte) error {
	if msgId > p.appendLastId ||
		p.appendFile == nil ||
		p.appendIndexFile == nil {

		if err := p.closeAppendFiles(); err != nil {
			return err
		}
		if err := p.createNextAppendFiles(msgId); err != nil {
			return err
		}
	}

	// write the message size and the message id 32bit and 64 bit
	sizeAndId := make([]byte, 12)
	binary.LittleEndian.PutUint32(sizeAndId, uint32(len(msg)))
	binary.LittleEndian.PutUint64(sizeAndId[4:], msgId)
	if _, err := p.appendFile.Write(sizeAndId); err != nil {
		return err
	}

	// write the message
	if _, err := p.appendFile.Write(msg); err != nil {
		return err
	}

	// write the index entry to the index file
	indexPosition := int64(12 * (msgId % MESSAGES_PER_FILE))
	messageOffset := p.appendFileWritePosition + uint64(len(sizeAndId))
	messageOffsetBuff := make([]byte, 12)
	binary.LittleEndian.PutUint64(messageOffsetBuff, messageOffset)
	binary.LittleEndian.PutUint32(messageOffsetBuff[8:], uint32(len(msg)))

	if _, err := p.appendIndexFile.WriteAt(messageOffsetBuff, indexPosition); err != nil {
		return err
	}

	p.appendFileWritePosition += uint64(len(sizeAndId) + len(msg))

	if msgId > p.maxMessageId {
		p.maxMessageId = msgId
	}
	return nil
}

// fetch a set of messages
func (p *MessagePartition) Fetch(req FetchRequest) {
	go func() {
		fetchList, err := p.calculateFetchList(req)
		if err != nil {
			req.ErrorCallback <- err
			return
		}

		err = p.fetchByFetchlist(fetchList, req.MessageC)
		if err != nil {
			req.ErrorCallback <- err
			return
		}
		close(req.MessageC)
	}()
}

// fetch the messages in the supplied fetchlist and send them to the channel
func (p *MessagePartition) fetchByFetchlist(fetchList []fetchEntry, messageC chan []byte) error {
	var fileId uint64
	var file *os.File
	var err error
	for _, f := range fetchList {
		// ensure, that we read on the correct file
		if file == nil || fileId != f.fileId {
			file, err = p.checkoutMessagefile(f.fileId)
			if err != nil {
				return err
			}
			defer p.releaseMessagefile(f.fileId, file)
			fileId = f.fileId
		}

		msg := make([]byte, f.size, f.size)
		_, err = file.ReadAt(msg, f.offset)
		if err != nil {
			return err
		}
		messageC <- msg
	}
	return nil
}

type fetchEntry struct {
	messageId uint64
	fileId    uint64
	offset    int64
	size      int
}

// returns a list of fetchEntry records for all message in the fetch request.
func (p *MessagePartition) calculateFetchList(req FetchRequest) ([]fetchEntry, error) {
	if req.Direction == 0 {
		req.Direction = 1
	}
	nextId := req.StartId
	result := make([]fetchEntry, 0, req.Count)
	var file *os.File
	var fileId uint64
	for len(result) < req.Count && nextId >= 0 {

		nextFileId := p.firstMessageIdForFile(nextId)

		// ensure, that we read on the correct file
		if file == nil || nextFileId != fileId {
			var err error
			file, err = p.checkoutIndexfile(nextFileId)
			if err != nil {
				if os.IsNotExist(err) {
					return result, nil
				}
				return nil, err
			}
			defer p.releaseIndexfile(fileId, file)
			fileId = nextFileId
		}

		indexPosition := int64(12 * (nextId % MESSAGES_PER_FILE))
		msgOffsetBuff := make([]byte, 12)

		if stat, err := file.Stat(); err != nil {
			return nil, err
		} else if indexPosition < stat.Size() {

			if _, err := file.ReadAt(msgOffsetBuff, indexPosition); err != nil {
				return nil, err
			}
			msgOffset := binary.LittleEndian.Uint64(msgOffsetBuff)
			msgSize := binary.LittleEndian.Uint32(msgOffsetBuff[8:])

			if msgOffset != uint64(0) { // only append, if the message exists
				result = append(result, fetchEntry{
					messageId: nextId,
					fileId:    fileId,
					offset:    int64(msgOffset),
					size:      int(msgSize),
				})
			}
		}

		nextId += uint64(req.Direction)
	}
	return result, nil
}

// Return a file handle to the message file with the supplied file id.
// The returned file handle may be shared for multiple go routinep.
func (p *MessagePartition) checkoutMessagefile(fileId uint64) (*os.File, error) {
	return os.Open(p.filenameByMessageId(fileId))
}

// Release a message file handle
func (p *MessagePartition) releaseMessagefile(fileId uint64, file *os.File) {
	file.Close()
}

// Return a file handle to the index file with the supplied file id.
// The returned file handle may be shared for multiple go routinep.
func (p *MessagePartition) checkoutIndexfile(fileId uint64) (*os.File, error) {
	return os.Open(p.indexFilenameByMessageId(fileId))
}

// Release an index file handle
func (p *MessagePartition) releaseIndexfile(fileId uint64, file *os.File) {
	file.Close()
}

func (p *MessagePartition) getSortedStartNumbersFromFiles() []uint64 {
	return []uint64{}
}

func (p *MessagePartition) firstMessageIdForFile(messageId uint64) uint64 {
	return messageId - messageId%MESSAGES_PER_FILE
}

func (p *MessagePartition) filenameByMessageId(messageId uint64) string {
	return filepath.Join(p.basedir, fmt.Sprintf("%s-%020d.msg", p.name, messageId))
}

func (p *MessagePartition) indexFilenameByMessageId(messageId uint64) string {
	return filepath.Join(p.basedir, fmt.Sprintf("%s-%020d.idx", p.name, messageId))
}