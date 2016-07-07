package store

import (
	"math/rand"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_SortedListSanity(t *testing.T) {

	a := assert.New(t)
	pq := newList(1000)

	generatedIds := make([]uint64, 0, 11)

	for i := 0; i < 11; i++ {
		msgID := uint64(rand.Intn(50))
		generatedIds = append(generatedIds, msgID)

		entry := &FetchEntry{
			size:      3,
			messageID: uint64(msgID),
			offset:    128,
		}
		pq.Insert(entry)
	}
	min := uint64(200)
	max := uint64(0)

	for _, id := range generatedIds {
		if max < id {
			max = id
		}
		if min > id {
			min = id
		}
		found, pos, _, foundEntry := pq.GetIndexEntryFromID(id)
		a.True(found)
		a.Equal(foundEntry.messageID, id)
		a.True(pos >= 0 && pos <= len(generatedIds))
	}

	logrus.WithField("generatedIds", generatedIds).Info("IdS")

	a.Equal(min, pq.Front().messageID)
	a.Equal(max, pq.Back().messageID)

	found, pos, bestIndex, foundEntry := pq.GetIndexEntryFromID(uint64(46))
	a.False(found, "Element should not be found since is a number greater than the random generated upper limit")
	a.Equal(pos, -1)
	a.Nil(foundEntry)
	logrus.WithField("bestIndexbestIndex", bestIndex).Info("bEST")

	a.Equal(pq.Front().messageID, pq.Get(0).messageID, "First element should contain the smallest element")
	a.Nil(pq.Get(-1), "Trying to get an invalid index will return nil")

	pq.Clear()
	a.Nil(pq.Front())
	a.Nil(pq.Back())

}
