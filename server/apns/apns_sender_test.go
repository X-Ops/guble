package apns

import (
	"github.com/golang/mock/gomock"
	"github.com/smancke/guble/protocol"
	"github.com/smancke/guble/server/router"
	"github.com/smancke/guble/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSender_ErrorBytes(t *testing.T) {
	a := assert.New(t)

	//given
	emptyBytes := []byte("")
	emptyPassword := ""
	cfg := Config{
		CertificateBytes:    &emptyBytes,
		CertificatePassword: &emptyPassword,
	}

	//when
	pusher, err := NewSender(cfg)

	// then
	a.Error(err)
	a.Nil(pusher)
}

func TestNewSender_ErrorFile(t *testing.T) {
	a := assert.New(t)

	//given
	wrongFilename := "."
	emptyPassword := ""
	cfg := Config{
		CertificateFileName: &wrongFilename,
		CertificatePassword: &emptyPassword,
	}

	//when
	pusher, err := NewSender(cfg)

	// then
	a.Error(err)
	a.Nil(pusher)
}

func TestSender_Send(t *testing.T) {
	_, finish := testutil.NewMockCtrl(t)
	defer finish()
	a := assert.New(t)

	// given
	routeParams := make(map[string]string)
	routeParams["device_id"] = "1234"
	routeConfig := router.RouteConfig{
		Path:        protocol.Path("path"),
		RouteParams: routeParams,
	}
	route := router.NewRoute(routeConfig)

	msg := &protocol.Message{
		Body: []byte("{}"),
	}

	mSubscriber := NewMockSubscriber(testutil.MockCtrl)
	mSubscriber.EXPECT().Route().Return(route).AnyTimes()

	mRequest := NewMockRequest(testutil.MockCtrl)
	mRequest.EXPECT().Subscriber().Return(mSubscriber).AnyTimes()
	mRequest.EXPECT().Message().Return(msg).AnyTimes()

	mPusher := NewMockPusher(testutil.MockCtrl)
	mPusher.EXPECT().Push(gomock.Any()).Return(nil, nil)

	// and
	s, err := NewSenderUsingPusher(mPusher, "com.myapp")
	a.NoError(err)

	// when
	rsp, err := s.Send(mRequest)

	// then
	a.NoError(err)
	a.Nil(rsp)
}
