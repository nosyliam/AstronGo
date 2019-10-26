package messagedirector

import (
	"astrongo/util"
	"sync"
)

var lock sync.Mutex

type Range struct {
	min util.Channel_t
	max util.Channel_t

	minBound util.Channel_t
	maxBound util.Channel_t
}

type RangeMap struct {
	rng         Range
	subscribers []*Subscriber
	intervals   map[Range]*Subscriber
}

func NewRangeMap(rng Range) *RangeMap {
	rm := &RangeMap{rng: rng}
	rm.intervals = make(map[Range]*Subscriber, 0)
	return rm
}

func (r *RangeMap) Send(ch util.Channel_t, dg util.Datagram) {
	// Subscribers of the original parent interval should receive
	if r.rng.min <= ch && r.rng.max >= ch {
		for sub := range r.subscribers {
		}
	}
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
	subscriptions map[util.Channel_t]<-chan interface{}

	// Ranges maps a range to a RangeMap, a structure which nests ranges
	ranges map[Range]*RangeMap
}

func (s *Subscriber) Subscribed(ch util.Channel_t) bool {
	for _, c := range s.channels {
		if c == ch {
			return true
		}
	}

	for _, rng := range s.ranges {
		if rng.min <= ch && rng.max >= ch {
			return true
		}
	}

	return false
}

func (c *ChannelMap) init() {
	c.subscriptions = make(map[util.Channel_t]<-chan interface{}, 0)
	c.ranges = make(map[Range]*RangeMap, 0)
}

func (c *ChannelMap) SubscribeRange(p *Subscriber, rng Range) {
	lock.Lock()
	defer lock.Unlock()

	p.ranges = append(p.ranges, rng)
	// Iterate through all channels that are ch>rng.Min && ch<rng.Max and push our subscriber.
	for ch, _ := range c.subscriptions {
		if ch > rng.min && ch < rng.max {
			c.SubscribeChannel(p, ch)
		}
	}
}

// Unsubscribing from a channel is a little more complicated; we must go through all of the ranges that contain
//  our subscriber and calculate new ranges so that there is no overlap with channels that are supposed to 'go silent'
func (c *ChannelMap) UnsubscribeRange(p *Subscriber, rng Range) {
	lock.Lock()
	defer lock.Unlock()

	for n, rng := range p.ranges {

	}
}

func (c *ChannelMap) SubscribeChannel(p *Subscriber, ch util.Channel_t) {
	lock.Lock()
	defer lock.Unlock()

	// Redundant, but we cannot check for range subscriptions when subscribing to a single channel
	for _, c := range p.channels {
		if c == ch {
			return
		}
	}

	p.channels = append(p.channels, ch)
	if _, ok := c.subscriptions[ch]; !ok {
		rdchan := make(chan interface{})
		go channelRoutine(rdchan)
		rdchan <- *p
		c.subscriptions[ch] = rdchan
	}
}

func (c *ChannelMap) Channel(ch util.Channel_t) <-chan interface{} {
	if chn, ok := c.subscriptions[ch]; !ok {
		return chn
	} else {
		// Default to range lookup
		rdchan := make(chan interface{})
		go func() {
			if dg, ok := (<-rdchan).(util.Datagram); ok {
				for rng, rnmap := range c.ranges {
					if rng.minBound <= ch && rng.maxBound >= ch {
						rnmap.Send(ch, dg)
					}
				}
			}
		}()
		return rdchan
	}

}

// channelRoutine implements a goroutine which continually reads a given chan for datagrams or subscriber objects.
//  When given a datagram, it assumes the channel associated with the routine is a receiver and will route the
//  the datagram to all of it's subscribers. When given a subscriber, it will append the object to it's subscribers
//  list; however, if the subscriber is inactive (denoted by subscriber.active) it assumes that a removal operation
//  operation is taking place and will attempt to a remove the subscriber from its subscribers list.
func channelRoutine(buf <-chan interface{}) {
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
					// Close dead channel
					if len(subscribers) == 0 {
						return
					}
				}
			case util.Datagram:
				for _, sub := range subscribers {
					sub.participant.HandleDatagram(data)
				}
			}
		}
	}
}
