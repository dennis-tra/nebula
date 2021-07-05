package service

import (
	"sync"
	"testing"
)

func TestNewService_lifecycle(t *testing.T) {
	s := New("test")

	s.ServiceStopped()
	s.ServiceStarted()
	s.ServiceStopped()
	s.ServiceStopped()
}

func TestNewService_shutdown(t *testing.T) {
	s := New("test")

	s.ServiceStarted()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		s.Shutdown()
		wg.Done()
	}()
	go s.ServiceStopped()
	wg.Wait()
}

func TestNewService_contexts_stopped(t *testing.T) {
	s := New("test")
	s.ServiceStarted()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-s.ServiceContext().Done()
			wg.Done()
		}()
	}
	go s.ServiceStopped()
	wg.Wait()
}

func TestNewService_contexts_shutdown(t *testing.T) {
	s := New("test")
	s.ServiceStarted()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-s.ServiceContext().Done()
			wg.Done()
		}()
	}
	go s.Shutdown()
	wg.Wait()
}

func TestNewService_restart(t *testing.T) {
	s := New("test")
	s.ServiceStarted()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			<-s.ServiceContext().Done()
			wg.Done()
		}()
	}
	wg.Add(1)
	go func() {
		s.Shutdown()
		wg.Done()
	}()
	go s.ServiceStopped()
	wg.Wait()
}

func TestService_SigDone(t *testing.T) {
	s := New("test")
	s.ServiceStarted()
	s.ServiceStopped()
	<-s.SigDone()
}

func TestService_SigShutdown(t *testing.T) {
	s := New("test")
	s.ServiceStarted()
	go s.Shutdown()
	<-s.SigShutdown()
}
