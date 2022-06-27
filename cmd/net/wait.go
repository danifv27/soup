package main

import "time"

type WaitFlags struct {
	Protocol  string        `enum:"tcp,udp" help:"Protocol to be used." default:"tcp" env:"SOUP_NET_PROTOCOL" prefix:"net.wait."`
	Addresses []string      `help:"address:port(,address:port,address:port,...)" env:"SOUP_NET_ADDRESSES" prefix:"net.wait."`
	Deadline  time.Duration `help:"deadline" env:"SOUP_NET_DEADLINE" prefix:"net.wait."`
	Wait      time.Duration `help:"delay of single requests" env:"SOUP_NET_WAIT" prefix:"net.wait."`
	Break     time.Duration `help:"break between requests" env:"SOUP_NET_BREAK" prefix:"net.wait."`
	Packet    string        `help:"UDP packet to be sent" env:"SOUP_NET_BREAK" prefix:"net.wait."`
}

// type WaitSubCmd struct {
// 	Flags WaitFlags `embed:"" prefix:"net.wait."`
// }

func (cmd *WaitCmd) Run(cli *CLI) error {
	// var err error
	// var apps application.Applications
	// var stopCh chan struct{}

	// if apps, err = initializeDiffCmd(cli, f); err != nil {
	// 	return fmt.Errorf("Run: %w", err)
	// }

	// wgLoop := &sync.WaitGroup{}

	// h := signals.NewSignalHandler([]os.Signal{syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM}, apps.LoggerService)
	// h.SetRunFunc(func() error {
	// 	wgLoop.Add(1)
	// 	stopCh = make(chan struct{})

	// 	apps.Informer.Start(stopCh)
	// 	wgLoop.Wait()

	// 	return nil
	// })
	// h.SetShutdownFunc(func(s os.Signal) error {
	// 	apps.LoggerService.WithFields(logger.Fields{
	// 		"signal": s,
	// 	}).Debug("Diff cmd shutting down")
	// 	close(stopCh)
	// 	wgLoop.Done()
	// 	event := audit.Event{
	// 		Action:  "DiffShutdown",
	// 		Actor:   "system",
	// 		Message: "diff command shutdown",
	// 	}
	// 	return apps.Auditer.Audit(&event)
	// })

	// ports := infrastructure.NewPorts(apps, &h)
	// ports.Actuators.SetActuatorRoot(cli.Diff.Actuator.Root)
	// wg := &sync.WaitGroup{}
	// ports.Actuators.Start(cli.Diff.Actuator.Address, wg, false, cli.Audit.Enable, "")
	// ports.MainLoop.Exec(wg, "diff cmd")
	// wg.Wait()

	return nil
}
