package messagedirector

import (
	. "astrongo/util"
	"sync"
)

var lock sync.Mutex
var channelMap *ChannelMap

type Range struct {
	Min Channel_t
	Max Channel_t
}

func (r Range) Size() Channel_t {
	return r.Max - r.Min
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
func (r *RangeMap) Split(rng Range, hi Channel_t, mid Channel_t, lo Channel_t, forward bool) Range {
	irng := r.intervals[rng]
	rnglo := Range{lo, mid}
	rnghi := Range{mid + 1, hi}

	for _, sub := range irng {
		intSubs := r.intervalSubs[sub]
		idx := 0
		for _, srng := range intSubs {
			if srng != rng && srng != rnglo && srng != rnghi {
				intSubs[idx] = srng
				idx++
			}
		}
		intSubs = intSubs[:idx]
		intSubs = addRange(intSubs, rnghi)
		intSubs = addRange(intSubs, rnglo)
		r.intervalSubs[sub] = intSubs
		sub.ranges = intSubs
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

func addSub(slice []*Subscriber, s *Subscriber) []*Subscriber {
	for _, sub := range slice {
		if sub == s {
			return slice
		}
	}
	return append(slice, s)
}

func addRange(slice []Range, r Range) []Range {
	for _, rng := range slice {
		if rng == r {
			return slice
		}
	}
	return append(slice, r)
}

func (r *RangeMap) Add(rng Range, sub *Subscriber) {
	lock.Lock()
	r.add(rng, sub)
	MD.AddRange(rng.Min, rng.Max)
	lock.Unlock()
}

func (r *RangeMap) add(rng Range, sub *Subscriber) {
	for erng, _ := range r.intervals {
		if rng == erng {
			break
		}

		// {xxxxxx[========]xxxxxxxx}
		if erng.Min > rng.Min && erng.Max < rng.Max {
			r.intervals[erng] = addSub(r.intervals[erng], sub)
			r.intervalSubs[sub] = addRange(r.intervalSubs[sub], erng)
			r.add(Range{rng.Min, erng.Min - 1}, sub)
			r.add(Range{erng.Max + 1, rng.Max}, sub)
			return
		}

		// [======{xxxxxxxx}========]
		if erng.Min < rng.Min && erng.Max > rng.Max {
			rng1 := r.Split(erng, erng.Max, rng.Min-1, erng.Min, true)
			rng2 := r.Split(rng1, rng1.Max, rng.Max, rng1.Min, false)
			r.intervals[rng2] = addSub(r.intervals[rng2], sub)
			r.intervalSubs[sub] = addRange(r.intervalSubs[sub], rng2)
			return
		}

		// [=============={xxxx]xxxx}
		if rng.Min >= erng.Min && rng.Min <= erng.Max && rng.Max > erng.Max {
			if rng.Min == erng.Min {
				r.intervals[erng] = addSub(r.intervals[erng], sub)
				r.intervalSubs[sub] = addRange(r.intervalSubs[sub], erng)
				r.add(Range{erng.Max + 1, rng.Max}, sub)
				return
			}
			nrng := r.Split(erng, erng.Max, rng.Min-1, erng.Min, true)
			r.intervals[nrng] = addSub(r.intervals[nrng], sub)
			r.intervalSubs[sub] = addRange(r.intervalSubs[sub], nrng)
			r.add(Range{erng.Max + 1, rng.Max}, sub)
			return
		}

		// {xxxx[xxxx}==============]
		if rng.Max >= erng.Min && rng.Max <= erng.Max && rng.Min < erng.Min {
			if rng.Max == erng.Max {
				r.intervals[erng] = addSub(r.intervals[erng], sub)
				r.intervalSubs[sub] = addRange(r.intervalSubs[sub], erng)
				r.add(Range{rng.Min, erng.Min - 1}, sub)
				return
			}
			nrng := r.Split(erng, erng.Max, rng.Max, erng.Min, false)
			r.intervals[nrng] = addSub(r.intervals[nrng], sub)
			r.intervalSubs[sub] = addRange(r.intervalSubs[sub], nrng)
			r.add(Range{rng.Min, erng.Min - 1}, sub)
			return
		}
	}

	r.intervals[rng] = addSub(r.intervals[rng], sub)
	r.intervalSubs[sub] = addRange(r.intervalSubs[sub], rng)
	return
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
		r.intervalSubs[p] = rng
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
		r.intervals[int] = rng
	}
}

func (r *RangeMap) Remove(rng Range, sub *Subscriber) {
	lock.Lock()
	r.remove(rng, sub)
	MD.RemoveRange(rng.Min, rng.Max)
	lock.Unlock()
}

func (r *RangeMap) remove(rng Range, sub *Subscriber) {
	for erng, _ := range r.intervals {
		// {xxxxxx[========]xxxxxxxx}
		if erng.Min >= rng.Min && erng.Max <= rng.Max {
			r.removeSubInterval(sub, erng)
			r.removeIntervalSub(erng, sub)
			if erng == rng {
				return
			}
			r.remove(Range{rng.Min, erng.Min - 1}, sub)
			r.remove(Range{erng.Max + 1, rng.Max}, sub)
			return
		}

		// [======{xxxxxxxx}========]
		if erng.Min < rng.Min && erng.Max > rng.Max {
			rng1 := r.Split(erng, erng.Max, rng.Min-1, erng.Min, true)
			rng2 := r.Split(rng1, rng1.Max, rng.Max, rng1.Min, false)
			r.removeSubInterval(sub, rng2)
			r.removeIntervalSub(rng2, sub)
			return
		}

		// [=============={xxxx]xxxx}
		if rng.Min >= erng.Min && rng.Min <= erng.Max && rng.Max > erng.Max {
			if rng.Min == erng.Min {
				r.removeSubInterval(sub, erng)
				r.removeIntervalSub(erng, sub)
				r.remove(Range{erng.Max + 1, rng.Max}, sub)
				return
			}
			nrng := r.Split(erng, erng.Max, rng.Min-1, erng.Min, true)
			r.removeSubInterval(sub, nrng)
			r.removeIntervalSub(nrng, sub)
			r.remove(Range{erng.Max + 1, rng.Max}, sub)
			return
		}

		// {xxxx[xxxx}==============]
		if rng.Max >= erng.Min && rng.Max <= erng.Max && rng.Min < erng.Min {
			if rng.Max == erng.Max {
				r.removeSubInterval(sub, erng)
				r.removeIntervalSub(erng, sub)
				r.remove(Range{rng.Min, erng.Min - 1}, sub)
				return
			}
			nrng := r.Split(erng, erng.Max, rng.Max, erng.Min, false)
			r.removeSubInterval(sub, nrng)
			r.removeIntervalSub(nrng, sub)
			r.remove(Range{rng.Min, erng.Min - 1}, sub)
			return
		}
	}
}

func (r *RangeMap) Send(ch Channel_t, dgi *DatagramIterator) {
	lock.Lock()
	defer lock.Unlock()

	for rng, subs := range r.intervals {
		if rng.Min <= ch && rng.Max >= ch {
			for _, sub := range subs {
				go sub.participant.HandleDatagram(*dgi.Dg, dgi)
			}
		}
	}
}

// Each MD participant is represented as a subscriber within the MD; when a participant desires to listen to
//  a DO (a "channel") the channel map will store it's ID in the participant's unique object.
type Subscriber struct {
	participant MDParticipant

	channels []Channel_t
	ranges   []Range

	active bool
}

type ChannelMap struct {
	// Subscriptions map channels to go channels which accepts datagram or Subscriber objects
	subscriptions sync.Map

	// Ranges points to a RangeMap singularity
	ranges *RangeMap
}

func (s *Subscriber) Subscribed(ch Channel_t) bool {
	for _, c := range s.channels {
		if c == ch {
			return true
		}
	}

	for _, rng := range s.ranges {
		if rng.Min <= ch && rng.Max >= ch {
			return true
		}
	}

	return false
}

func (c *ChannelMap) init() {
	c.ranges = NewRangeMap()
}

func (c *ChannelMap) SubscribeRange(p *Subscriber, rng Range) {
	// Remove single-channel subscriptions; we can't risk data being sent twice
	for _, ch := range p.channels {
		if rng.Min <= ch && rng.Max >= ch {
			c.UnsubscribeChannel(p, ch)
		}
	}

	c.ranges.Add(rng, p)
	p.ranges = c.ranges.Ranges(p)
}

func (c *ChannelMap) UnsubscribeRange(p *Subscriber, rng Range) {
	c.ranges.Remove(rng, p)
	p.ranges = c.ranges.Ranges(p)

}

func (c *ChannelMap) UnsubscribeChannel(p *Subscriber, ch Channel_t) {
	if chn, ok := c.subscriptions.Load(ch); ok {
		chn := chn.(chan interface{})
		p.active = false
		cpy := Subscriber(*p)
		chn <- &cpy
		p.active = true

		idx := 0
		for _, c := range p.channels {
			if c != ch {
				p.channels[idx] = c
				idx++
			}
		}
		p.channels = p.channels[:idx]
		MD.RemoveChannel(ch)
	} else {
		c.ranges.Remove(Range{ch, ch}, p)
	}
}

func (c *ChannelMap) UnsubscribeAll(p *Subscriber) {
	c.UnsubscribeRange(p, Range{0, CHANNEL_MAX})
	for _, ch := range p.channels {
		c.UnsubscribeChannel(p, ch)
	}
}

func (c *ChannelMap) SubscribeChannel(p *Subscriber, ch Channel_t) {
	if p.Subscribed(ch) {
		return
	}

	p.channels = append(p.channels, ch)
	if chn, ok := c.subscriptions.Load(ch); !ok {
		rdchan := make(chan interface{})
		go channelRoutine(rdchan, ch)
		rdchan <- p
		c.subscriptions.Store(ch, rdchan)
	} else {
		chn := chn.(chan interface{})
		chn <- p
	}
}

func (c *ChannelMap) Channel(ch Channel_t) chan interface{} {
	if chn, ok := c.subscriptions.Load(ch); ok {
		chn := chn.(chan interface{})
		return chn
	} else {
		// Default to range lookup
		rdchan := make(chan interface{})
		go func() {
			if dg, ok := (<-rdchan).(*DatagramIterator); ok {
				c.ranges.Send(ch, dg)
			}
		}()
		return rdchan
	}

}

// channelRoutine implements a goroutine which continually reads a given chan for datagram or subscriber objects.
//  When given a datagram, it uses channel associated with the routine is a receiver and will route the
//  the datagram to all of its subscribers. When given a subscriber, it will append the object to its subscribers
//  list; however, if the subscriber is inactive (denoted by subscriber.active) it assumes that a removal operation
//  operation is taking place and will attempt to remove it from the subscribers list.
func channelRoutine(buf chan interface{}, ch Channel_t) {
	var subscribers []*Subscriber
	MD.AddChannel(ch)
	for {
		select {
		case v, ok := <-buf:
			if !ok {
				break
			}

			switch data := v.(type) {
			case *Subscriber:
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
						channelMap.subscriptions.Delete(ch)
						MD.RemoveChannel(ch)
						return
					}
				}
			case *DatagramIterator:
				channelMap.ranges.Send(ch, data)

				for _, sub := range subscribers {
					go sub.participant.HandleDatagram(*data.Dg, data)
				}
			}
		}
	}
}

func init() {
	channelMap = &ChannelMap{}
	channelMap.init()
}
