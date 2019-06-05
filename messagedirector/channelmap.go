package messagedirector

import (
	"astrongo/util"
	"sync"
)

type Range struct {
	min util.Channel_t
	max util.Channel_t
}

// Each MD participant is represented as a subscriber within the MD; when a participant desires to listen to
//  a DO (a "channel") the channel map will store it's ID in the participant's unique object.
type Subscriber struct {
	participant MDParticipant

	channels []util.Channel_t
	ranges   []Range

	active bool
}

type ChannelMap struct {
	// Subscriptions map channels to a chan which accepts datagram or Subscriber objects
	subscriptions sync.Map

	// Ranges maps a range to a chan which accepts datagram or Subscriber objects
	ranges sync.Map
}

func (s *Subscriber) Subscribed(ch util.Channel_t) bool {
	for _, c := range s.channels {
		if c == ch {
			return true
		}
	}

	for _, rng := range s.ranges {
		if rng.min < ch && rng.max > ch {
			return true
		}
	}

	return false
}

func (c *ChannelMap) SubscribeRange(p *Subscriber, rng Range) {

}

func (c *ChannelMap) SubscribeChannel(p *Subscriber, ch util.Channel_t) {
	if p.Subscribed(ch) {
		return
	}

	p.channels = append(p.channels, ch)
	if _, ok := c.subscriptions.Load(ch); !ok {
		rdchan := make(chan interface{})
		go channelRoutine(rdchan, ch)
		rdchan <- *p
		c.subscriptions.Store(ch, rdchan)
	}

}

// channelRoutine implements a goroutine which continually reads a given chan for datagram and subscriber objects.
//  When given a datagram, it assumes the channel associated with the routine is a receiver and will route the
//  the datagram to all of it's subscribers. When given a subscriber, it will append the object to it's subscribers
//  list; however, if the subscriber is inactive (denoted by subscriber.active) it assumes that a removal operation
//  operation is taking place and will attempt to a remove the subscriber from it's subscribers list.
func channelRoutine(buf <-chan interface{}, channel util.Channel_t) {
	var subscribers []Subscriber
	for {
		select {
		case v, ok := <-buf:
			if !ok {
				break
			}

			switch data := v.(type) {
			case Subscriber:
				if data.active {
					subscribers = append(subscribers, data)
				} else {
					idx := 0
					for _, sub := range subscribers {
						if sub.participant != data.participant {
							subscribers[idx] = sub
							idx++
						}
					}
					subscribers = subscribers[:idx]
				}
			case util.Datagram:
				for _, sub := range subscribers {
					sub.participant.HandleDatagram(data)
				}
			}
		}
	}

}
