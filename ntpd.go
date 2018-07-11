package gontpd

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/beevik/ntp"
)

var errNoMedian = errors.New("no median found")

type NTPd struct {
	template []byte

	cfg *Config

	peerList []*peer
	stat     *statistic
	sleep    time.Duration

	delay time.Duration
	disp  time.Duration
}

func New(cfg *Config) (d *NTPd) {

	d = &NTPd{cfg: cfg,
		template: newTemplate()}
	if cfg.Metric != "" {
		d.stat = newStatistic(cfg)
	}
	return d
}

func (d *NTPd) Run() (err error) {

	err = d.init()
	if err != nil {
		return
	}

	d.poll()
	median := d.find()
	if median == nil {
		err = errNoMedian
		return
	}
	err = syncClock(median.resp.ClockOffset, 0,
		d.cfg.ForceUpdate)
	if err != nil {
		log.Println("sync err:", err, " offset:", median.resp.ClockOffset)
		return
	}

	go d.listen()

	for {
		log.Printf("wait for %s", d.sleep.String())
		time.Sleep(d.sleep)

		d.poll()
		median = d.find()
		if median == nil {
			log.Println(errNoMedian)
			d.sleep = time.Second * 10
			continue
		}

		err = syncClock(median.resp.ClockOffset,
			uint8(median.resp.Leap), d.cfg.ForceUpdate)
		if err != nil {
			return
		}

		d.setTemplate(median)
		d.updateState(median)

		if absDuration(median.resp.ClockOffset) < time.Millisecond*480 {
			poll := median.peer.trustLevel
			if poll > d.cfg.MaxPoll {
				poll = d.cfg.MaxPoll
			}
			if poll < d.cfg.MinPoll {
				poll = d.cfg.MinPoll
			}

			s := math.Exp2(float64(poll))
			d.sleep = time.Duration(s) * time.Second
			if d.sleep < time.Second*10 {
				d.sleep = time.Second * 10
			}
			continue
		}
		d.sleep = 10 * time.Second
		for i := 0; i < len(d.peerList); i++ {
			d.peerList[i].trustLevel = 1
		}
	}
}

func (d *NTPd) updateState(op *offsetPeer) {
	if d.stat != nil {
		d.stat.delayGauge.Set(op.resp.RootDelay.Seconds())
		d.stat.offsetGauge.Set(op.resp.ClockOffset.Seconds())
		d.stat.dispGauge.Set(op.resp.RootDispersion.Seconds())
	}
}

func (d *NTPd) init() (err error) {
	for _, addr := range d.cfg.PeerList {
		p := newPeer(addr)
		if p == nil {
			log.Print("peer:%s init failed", addr)
		}
		d.peerList = append(d.peerList, p)
	}

	if len(d.peerList) == 0 {
		err = fmt.Errorf("no available peer, tried: %v", d.cfg.PeerList)
	}
	return
}

func (d *NTPd) poll() {
	var wg sync.WaitGroup
	wg.Add(len(d.peerList))
	for _, p := range d.peerList {
		go func(p *peer) {
			p.update()
			wg.Done()
		}(p)
	}
	wg.Wait()

	goodCount := 0
	for _, p := range d.peerList {
		if p.good {
			goodCount += 1
		}
	}
	if goodCount < 3 {
		log.Print("not enough good peers, but continue")
	}
}

type offsetPeer struct {
	peer *peer
	resp *ntp.Response
}

func (d *NTPd) find() (op *offsetPeer) {

	tmp := []*offsetPeer{}
	for _, p := range d.peerList {
		if !p.good {
			continue
		}

		for _, resp := range p.reply {
			if resp.Stratum >= invalidStratum {
				continue
			}
			tmp = append(tmp, &offsetPeer{p, resp})
		}
	}

	if len(tmp) == 0 {
		return
	}
	sort.Sort(byOffset(tmp))
	if debug {
		for _, p := range tmp {
			fmt.Printf("%s:%s,", p.peer.addr, p.resp.ClockOffset)
		}
		fmt.Print("\n")
	}

	if len(tmp) < goodFilter {
		return
	}

	op = tmp[len(tmp)/2]
	return
}

type byOffset []*offsetPeer

func (b byOffset) Len() int {
	return len(b)
}

func (b byOffset) Less(i, j int) bool {
	return b[i].resp.ClockOffset < b[j].resp.ClockOffset
}

func (b byOffset) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
