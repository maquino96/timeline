package scheduler

import (
	"log"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/maquino96/timeline/internal/adapters"
	"github.com/maquino96/timeline/internal/models"
	"github.com/maquino96/timeline/internal/store"
	"github.com/maquino96/timeline/internal/topic"
)

const domainMinInterval = 90 * time.Second
const rateLimitCooldown = 10 * time.Minute

type Scheduler struct {
	store           *store.Store
	registry        *adapters.Registry
	engine          *topic.Engine
	mu              sync.Mutex
	stopCh          chan struct{}
	running         map[int64]bool
	lastPoll        map[int64]time.Time
	domainLastPoll  map[string]time.Time
	domainCooldown  map[string]time.Time
}

func New(s *store.Store, r *adapters.Registry, e *topic.Engine) *Scheduler {
	return &Scheduler{
		store:          s,
		registry:       r,
		engine:         e,
		stopCh:         make(chan struct{}),
		running:        make(map[int64]bool),
		lastPoll:       make(map[int64]time.Time),
		domainLastPoll: make(map[string]time.Time),
		domainCooldown: make(map[string]time.Time),
	}
}

func (sc *Scheduler) Start() {
	go sc.loop()
}

func (sc *Scheduler) Stop() {
	close(sc.stopCh)
}

func (sc *Scheduler) loop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sc.stopCh:
			return
		case <-ticker.C:
			sc.pollDue()
		}
	}
}

func (sc *Scheduler) pollDue() {
	sources, err := sc.store.GetEnabledSources()
	if err != nil {
		log.Printf("scheduler: get sources: %v", err)
		return
	}

	sort.Slice(sources, func(i, j int) bool {
		return sc.lastPoll[sources[i].ID].Before(sc.lastPoll[sources[j].ID])
	})

	for _, src := range sources {
		sc.mu.Lock()
		last, exists := sc.lastPoll[src.ID]
		if exists && time.Since(last) < time.Duration(src.Interval)*time.Second {
			sc.mu.Unlock()
			continue
		}
		if sc.running[src.ID] {
			sc.mu.Unlock()
			continue
		}
		domain := extractDomain(src)
		if domain != "" {
			if cooldown, ok := sc.domainCooldown[domain]; ok && time.Now().Before(cooldown) {
				sc.mu.Unlock()
				continue
			}
			lastDomain, domainExists := sc.domainLastPoll[domain]
			if domainExists && time.Since(lastDomain) < domainMinInterval {
				sc.mu.Unlock()
				continue
			}
			sc.domainLastPoll[domain] = time.Now()
		}
		sc.running[src.ID] = true
		sc.lastPoll[src.ID] = time.Now()
		sc.domainLastPoll[domain] = time.Now()
		sc.mu.Unlock()

		go sc.fetchSource(src)
	}
}

func (sc *Scheduler) fetchSource(source models.Source) {
	defer func() {
		sc.mu.Lock()
		delete(sc.running, source.ID)
		sc.mu.Unlock()
	}()

	items, err := sc.registry.Fetch(source)
	if err != nil {
		log.Printf("scheduler: fetch %s: %v", source.Name, err)
		if strings.Contains(err.Error(), "status 429") {
			domain := extractDomain(source)
			sc.mu.Lock()
			sc.domainCooldown[domain] = time.Now().Add(rateLimitCooldown)
			log.Printf("scheduler: rate limited by %s, cooling down for %v", domain, rateLimitCooldown)
			sc.mu.Unlock()
		}
		return
	}

	topics, err := sc.store.GetEnabledTopics()
	if err != nil {
		log.Printf("scheduler: get topics: %v", err)
	}

	for i := range items {
		if err := sc.store.UpsertItem(&items[i]); err != nil {
			log.Printf("scheduler: upsert item: %v", err)
			continue
		}

		for _, topicID := range sc.engine.Match(&items[i], topics) {
			if err := sc.store.AddItemTopic(items[i].ID, topicID); err != nil {
				log.Printf("scheduler: add item topic: %v", err)
			}
		}
	}

	if len(items) > 0 {
		log.Printf("scheduler: fetched %d items from %s (%s)", len(items), source.Name, source.Type)
	}
}

func (sc *Scheduler) PollSourceNow(sourceID int64) error {
	sources, err := sc.store.GetSources()
	if err != nil {
		return err
	}
	for _, src := range sources {
		if src.ID == sourceID {
			sc.mu.Lock()
			sc.lastPoll[sourceID] = time.Now()
			sc.mu.Unlock()

			items, err := sc.registry.Fetch(src)
			if err != nil {
				return err
			}
			topics, _ := sc.store.GetEnabledTopics()
			for i := range items {
				sc.store.UpsertItem(&items[i])
				for _, topicID := range sc.engine.Match(&items[i], topics) {
					sc.store.AddItemTopic(items[i].ID, topicID)
				}
			}
			return nil
		}
	}
	return nil
}

func extractDomain(source models.Source) string {
	if source.Type == models.SourceSECEDGAR {
		return "www.sec.gov"
	}
	u, err := url.Parse(source.URL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}
