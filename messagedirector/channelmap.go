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
//  a DO (a "channel") the channel map will store the ID in the participant's unique object. This subscription
//  is also recorded in the MD's subscription map, inserting the subscriber into the channel's subscribed channels list.
type Subscriber struct {
	participant MDParticipant

	channels []util.Channel_t
	ranges   []Range
}

type ChannelMap struct {
	// Subscriptions map channels to a list of their subscribers
	subscriptions map[util.Channel_t][]Subscriber

	// Ranges maps a range to a list of it's subscribers
	ranges map[Range][]Subscriber
	lock   sync.Mutex
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

// Lookup asynchronously searches for the set of channels within the channel and
//  range maps and returns the subscription lists of the results
func (c *ChannelMap) Lookup(channels []util.Channel_t) []MDParticipant {
	var participants []MDParticipant

	c.lock.Lock()
	defer c.lock.Unlock()

	ps := make(chan MDParticipant)
	finish := make(chan int)

	go func() {
		for _, ch := range channels {
			// Search if the channel possesses a subscription list; if it does, append all of it
			if subs, ok := c.subscriptions[ch]; ok {
				for _, sub := range subs {
					ps <- sub.participant
				}
			}
		}
		finish <- 1
	}()

	go func() {
		for _, ch := range channels {
			// Search each range to see if it contains the channel; if it does, append it's subscription list
			for rng, subs := range c.ranges {
				if rng.min < ch && rng.max > ch {
					for _, sub := range subs {
						ps <- sub.participant
					}
				}
			}

		}
		finish <- 1
	}()

	<-finish
	<-finish
	close(ps)

	for sub := range ps {
		participants = append(participants, sub)
	}
	return participants
}

func (c *ChannelMap) SubscribeChannel(p *Subscriber, ch util.Channel_t) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if p.Subscribed(ch) {
		return
	}

	p.channels = append(p.channels, ch)
	if subs, ok := c.subscriptions[ch]; ok {
		// TODO: a
	}
}
