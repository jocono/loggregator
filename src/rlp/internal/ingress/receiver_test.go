package ingress_test

import (
	"errors"
	"fmt"
	"plumbing"

	v2 "plumbing/v2"
	"rlp/internal/ingress"

	"golang.org/x/net/context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Receiver", func() {
	var (
		spyConverter  *SpyEnvelopeConverter
		spySubscriber *SpySubscriber
		receiver      *ingress.Receiver
	)

	BeforeEach(func() {
		spyConverter = &SpyEnvelopeConverter{}
		spySubscriber = &SpySubscriber{
			recv: func() ([]byte, error) {
				return []byte("something"), nil
			},
		}
		receiver = ingress.NewReceiver(spyConverter, ingress.NewRequestConverter(), spySubscriber)
	})

	It("streams converted data", func() {
		expectedEnv := &v2.Envelope{Timestamp: 1}
		spyConverter.envelope = expectedEnv

		req := &v2.EgressRequest{}
		receiver, err := receiver.Receive(context.Background(), req)
		Expect(err).ToNot(HaveOccurred())

		env, err := receiver()
		Expect(err).ToNot(HaveOccurred())
		Expect(env).To(Equal(expectedEnv))
	})

	It("subscribes to data", func() {
		req := &v2.EgressRequest{
			ShardId: "some-id",
			Filter: &v2.Filter{
				SourceId: "some-source-id",
				Message: &v2.Filter_Log{
					Log: &v2.LogFilter{},
				},
			},
		}
		expectedReq := &plumbing.SubscriptionRequest{
			ShardID: req.ShardId,
			Filter: &plumbing.Filter{
				AppID: "some-source-id",
				Message: &plumbing.Filter_Log{
					Log: &plumbing.LogFilter{},
				},
			},
		}
		receiver.Receive(context.Background(), req)

		Expect(spySubscriber.req).To(Equal(expectedReq))
	})

	It("converts the data", func() {
		req := &v2.EgressRequest{}
		receiver, err := receiver.Receive(context.Background(), req)
		Expect(err).ToNot(HaveOccurred())
		receiver()

		Expect(spyConverter.data).To(Equal([]byte("something")))
	})

	It("returns an error if the convert fails", func() {
		spyConverter.err = errors.New("some error")
		req := &v2.EgressRequest{}
		receiver, err := receiver.Receive(context.Background(), req)
		Expect(err).ToNot(HaveOccurred())
		_, err = receiver()

		Expect(err).To(HaveOccurred())
	})

	It("returns an error via the receiver", func() {
		spySubscriber.recv = func() ([]byte, error) {
			return nil, fmt.Errorf("some-error")
		}
		req := &v2.EgressRequest{}
		receiver, err := receiver.Receive(context.Background(), req)
		Expect(err).ToNot(HaveOccurred())

		_, err = receiver()
		Expect(err).To(HaveOccurred())
	})

	It("returns an error when the subscriber fails", func() {
		spySubscriber.err = errors.New("some error")
		req := &v2.EgressRequest{}
		_, err := receiver.Receive(context.Background(), req)
		Expect(err).To(HaveOccurred())
	})

})

type SpyEnvelopeConverter struct {
	data     []byte
	envelope *v2.Envelope
	err      error
}

func (s *SpyEnvelopeConverter) Convert(data []byte) (*v2.Envelope, error) {
	s.data = data
	return s.envelope, s.err
}

type SpySubscriber struct {
	req  *plumbing.SubscriptionRequest
	recv func() ([]byte, error)
	err  error
}

func (s *SpySubscriber) Subscribe(ctx context.Context, req *plumbing.SubscriptionRequest) (recv func() ([]byte, error), err error) {
	s.req = req
	return s.recv, s.err
}
