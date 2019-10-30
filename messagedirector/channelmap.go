package messagedirector

import (
	"astrongo/util"
	"sync"
)

var lock sync.Mutex
var channelMap *ChannelMap

type Range struct {
	min util.Channel_t
	max util.Channel_t

	minBound util.Channel_t
	maxBound util.Channel_t
}

type RangeMap struct {
	rng          Range
	subscribers  []*Subscriber
	intervals    map[Range][]*Subscriber
	intervalSubs map[*Subscriber][]Range
}

func NewRangeMap(rng Range) *RangeMap {
	rm := &RangeMap{rng: rng}
	rm.intervals = make(map[Range][]*Subscriber, 0)
	rm.intervalSubs = make(map[*Subscriber][]Range, 0)
	return rm
}

func (r *RangeMap) Split(rng Range, hi util.Channel_t, mid util.Channel_t, lo util.Channel_t, forward bool) Range {
	irng := r.intervals[rng]
	rnglo := Range{min: lo, max: mid}
	rnghi := Range{min: mid, max: hi}

	for _, sub := range irng {
		intSubs := r.intervalSubs[sub]
		idx := 0
		for _, srng := range intSubs {
			if srng != rng {
				intSubs[idx] = srng
				idx++
			}
		}
		intSubs = intSubs[:idx]
		intSubs = append(append(intSubs, rnglo), rnghi)
	}

	r.intervals[rnglo] = append(make([]*Subscriber, 0, len(irng)), irng...)
	r.intervals[rnghi] = append(make([]*Subscriber, 0, len(irng)), irng...)
	delete(r.intervals, rng)

	if forward {
		return rnghi
	} else {
		return rnglo
	}
}

func (r *RangeMap) Add(rng Range, sub *Subscriber) {
	if rng.min > r.rng.max || rng.max < r.rng.min {
		// This should never happen, but the range is invalid.
		panic(2)
	}

	if rng == r.rng {
		r.subscribers = append(r.subscribers, sub)
		return
	}

	// Check to see if the range is nested within or overlapping other ranges
	if rngs, ok := r.intervalSubs[sub]; ok {
		for _, erng := range rngs {
			// [======{xxxxxxxx}========]
			if erng.min <= rng.min && erng.max >= rng.max {
				// Nested range. We don't need to do anything besides add it
				break
			}

			// [=============={xxxx]xxxx}
			if erng.max < rng.max || erng.min > rng.min {
				nrng := r.Split(erng, erng.max, rng.min, erng.min, true)
				trng := Range{min: erng.max, max: rng.max}
				r.intervals[nrng] = append(r.intervals[nrng], sub)
				r.intervalSubs[sub] = append(r.intervalSubs[sub], trng)
				r.Add(trng, sub)
				return
			}

			// {xxxx[xxxx}==============]
			if erng.min > rng.min {
				nrng := r.Split(erng, erng.max, rng.max, erng.min, false)
				trng := Range{min: rng.min, max: erng.min}
				r.intervals[nrng] = append(r.intervals[nrng], sub)
				r.intervalSubs[sub] = append(r.intervalSubs[sub], trng)
				r.Add(trng, sub)
				return
			}
		}
	}

	r.intervals[rng] = append(r.intervals[rng], sub)
	r.intervalSubs[sub] = append(r.intervalSubs[sub], rng)
}

func (r *RangeMap) Send(ch util.Channel_t, dg util.Datagram) {
	// Subscribers of the original parent interval should receive the datagram
	if r.rng.min <= ch && r.rng.max >= ch {
		for _, sub := range r.subscribers {
			sub.participant.HandleDatagram(dg)
		}
	}

	for rng, subs := range r.intervals {
		if rng.min <= ch && rng.max >= ch {
			for _, sub := range subs {
				sub.participant.HandleDatagram(dg)
			}
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
	// Remove single-channel subscriptions; we can't risk data being sent twice
	for ch, _ := range c.subscriptions {
		if ch > rng.min && ch < rng.max {
			c.UnsubscribeChannel(p, ch)
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

func (c *ChannelMap) UnsubscribeChannel(p *Subscriber, ch util.Channel_t) {
	lock.Lock()
	defer lock.Unlock()
}

func (c *ChannelMap) SubscribeChannel(p *Subscriber, ch util.Channel_t) {
	lock.Lock()
	defer lock.Unlock()

	if p.Subscribed(ch) {
		return
	}

	p.channels = append(p.channels, ch)
	if _, ok := c.subscriptions[ch]; !ok {
		rdchan := make(chan interface{})
		go channelRoutine(rdchan, ch)
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
func channelRoutine(buf <-chan interface{}, ch util.Channel_t) {
	var subscribers []Subscriber
	var rngmap *RangeMap // Range mappings are immutable so we can cache them
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
				// We need to do this within the channel-subscription routine because we will never be able
				//  to kow about range subscriptions created before the routine was started
				if rngmap == nil {
					for rng, rnmap := range channelMap.ranges {
						if rng.minBound <= ch && rng.maxBound >= ch {
							rngmap = rnmap
						}
					}
				}
				if rngmap != nil {
					rngmap.Send(ch, data)
				}

				for _, sub := range subscribers {
					sub.participant.HandleDatagram(data)
				}
			}
		}
	}
}

func init() {
	channelMap = &ChannelMap{}
	channelMap.init()
}
