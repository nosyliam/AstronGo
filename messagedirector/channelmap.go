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
	subscribers  []*Subscriber
	intervals    map[Range][]*Subscriber
	intervalSubs map[*Subscriber][]Range
}

func NewRangeMap() *RangeMap {
	rm := &RangeMap{}
	rm.intervals = make(map[Range][]*Subscriber, 0)
	rm.intervalSubs = make(map[*Subscriber][]Range, 0)
	return rm
}

// Returns an empty range or a list of ranges a subscriber is subscribed to; should usually never return []
func (r *RangeMap) Ranges(p *Subscriber) []Range {
	if rngs, ok := r.intervalSubs[p]; ok {
		return rngs
	} else {
		return make([]Range, 0)
	}
}

// Splits a range; e.g. [================] => [=======][==========]
func (r *RangeMap) Split(rng Range, hi util.Channel_t, mid util.Channel_t, lo util.Channel_t, forward bool) Range {
	irng := r.intervals[rng]
	rnglo := Range{min: lo, max: mid}
	rnghi := Range{min: mid + 1, max: hi}

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
	// Check to see if the range is nested within or overlapping other ranges
	if rngs, ok := r.intervalSubs[sub]; ok {
		for _, erng := range rngs {
			// [======{xxxxxxxx}========]
			if erng.min <= rng.min && erng.max >= rng.max {

				return
			}

			// [=============={xxxx]xxxx}
			if erng.max < rng.max && rng.min < erng.max {
				nrng := r.Split(erng, erng.max, rng.min, erng.min, true)
				trng := Range{min: erng.max, max: rng.max}
				r.intervals[nrng] = append(r.intervals[nrng], sub)
				r.intervalSubs[sub] = append(r.intervalSubs[sub], trng)
				r.Add(trng, sub)
				return
			}

			// {xxxx[xxxx}==============]
			if erng.min > rng.min && rng.max > erng.min {
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

func (r *RangeMap) removeSubInterval(p *Subscriber, rmv Range) {
	if rng, ok := r.intervalSubs[p]; ok {
		idx := 0
		for _, v := range rng {
			if v != rmv {
				rng[idx] = v
				idx++
			}
		}
		rng = rng[:idx]
	}
}

func (r *RangeMap) removeIntervalSub(int Range, p *Subscriber) {
	if rng, ok := r.intervals[int]; ok {
		idx := 0
		for _, v := range rng {
			if v != p {
				rng[idx] = v
				idx++
			}
		}
		rng = rng[:idx]
	}
}

func (r *RangeMap) Remove(rng Range, sub *Subscriber) {
	if rngs, ok := r.intervalSubs[sub]; ok {
		for _, erng := range rngs {
			// [======{xxxxxxxx}========] => [=====][xxxxxxx][=======]
			if erng.min <= rng.min && erng.max >= rng.max {
				rng1 := r.Split(erng, erng.max, rng.min-1, erng.min, true)
				rng2 := r.Split(rng1, rng1.min, rng.max, rng1.max, false)
				r.removeSubInterval(sub, rng2)
				r.removeIntervalSub(rng2, sub)
				break
			}

			// [=============={xxxx]xxxx} => [===========][xxxx][????]
			if erng.max < rng.max && rng.min < erng.max {
				nrng := r.Split(erng, erng.max, rng.min-1, erng.min, true)
				trng := Range{min: erng.max + 1, max: rng.max}
				r.removeSubInterval(sub, nrng)
				r.removeIntervalSub(nrng, sub)
				r.Remove(trng, sub)
				return
			}

			// {xxxx[xxxx}==============] => [????][xxxx][=============]
			if erng.min > rng.min && rng.max > erng.min {
				nrng := r.Split(erng, erng.max, rng.max, erng.min, false)
				trng := Range{min: rng.min, max: erng.min - 1}
				r.removeSubInterval(sub, nrng)
				r.removeIntervalSub(nrng, sub)
				r.Remove(trng, sub)
				return
			}
		}
	}
}

func (r *RangeMap) Send(ch util.Channel_t, dg util.Datagram) {
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
	subscriptions map[util.Channel_t]chan interface{}

	// Ranges points to a RangeMap singularity
	ranges *RangeMap
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
	c.subscriptions = make(map[util.Channel_t]chan interface{}, 0)
	c.ranges = NewRangeMap()
}

func (c *ChannelMap) SubscribeRange(p *Subscriber, rng Range) {
	lock.Lock()
	defer lock.Unlock()

	// Remove single-channel subscriptions; we can't risk data being sent twice
	for ch, _ := range c.subscriptions {
		if rng.min <= ch && rng.max >= ch {
			c.UnsubscribeChannel(p, ch)
		}
	}

	c.ranges.Add(rng, p)
	p.ranges = c.ranges.Ranges(p)
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

	if chn, ok := c.subscriptions[ch]; ok {
		p.active = false
		chn <- *p
		p.active = true
	} else {

	}
}

func (c *ChannelMap) SubscribeChannel(p *Subscriber, ch util.Channel_t) {
	lock.Lock()
	defer lock.Unlock()

	if p.Subscribed(ch) {
		return
	}

	p.channels = append(p.channels, ch)
	if chn, ok := c.subscriptions[ch]; !ok {
		rdchan := make(chan interface{})
		go channelRoutine(rdchan, ch)
		rdchan <- *p
		c.subscriptions[ch] = rdchan
	} else {
		chn <- *p
	}
}

func (c *ChannelMap) Channel(ch util.Channel_t) chan interface{} {
	if chn, ok := c.subscriptions[ch]; !ok {
		return chn
	} else {
		// Default to range lookup
		rdchan := make(chan interface{})
		go func() {
			if dg, ok := (<-rdchan).(util.Datagram); ok {
				c.ranges.Send(ch, dg)
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
func channelRoutine(buf chan interface{}, ch util.Channel_t) {
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
				channelMap.ranges.Send(ch, data)

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
