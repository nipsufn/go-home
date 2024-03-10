package config

import "sync"

type BulbState struct {
	On          bool
	Brightness  uint8
	Temperature uint
	Color       string
}

type State struct {
	mu    sync.RWMutex
	bulbs map[string]BulbState
}

func (s *State) Set(name string, state BulbState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.bulbs[name] = state
}

func (s *State) SetOn(name string, on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tmp := s.bulbs[name]
	tmp.On = on

	s.bulbs[name] = tmp
}

func (s *State) SetBrightness(name string, brightness uint8) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tmp := s.bulbs[name]
	tmp.Brightness = brightness

	s.bulbs[name] = tmp
}

func (s *State) SetTemperature(name string, temperature uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tmp := s.bulbs[name]
	tmp.Temperature = temperature

	s.bulbs[name] = tmp
}

func (s *State) SetColor(name string, color string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tmp := s.bulbs[name]
	tmp.Color = color

	s.bulbs[name] = tmp
}

func (s *State) SetMasterState(on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tmp := s.bulbs[ConfigSingleton.Bulb.Master]
	tmp.On = on

	s.bulbs[ConfigSingleton.Bulb.Master] = tmp
}

func (s *State) Get(name string) BulbState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.bulbs[name]
}

func (s *State) GetMasterState() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.bulbs[ConfigSingleton.Bulb.Master].On
}

func (s *State) Init() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.bulbs = make(map[string]BulbState)
}
