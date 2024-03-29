package main

import (
	"astrongo/clientagent"
	"astrongo/core"
	"astrongo/dclass/dc"
	"astrongo/eventlogger"
	"astrongo/messagedirector"
	"astrongo/util"
	"fmt"
	"github.com/apex/log"
	"github.com/spf13/pflag"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
)

var mainLog *log.Entry

func init() {
	log.SetHandler(core.Log)
	log.SetLevel(log.DebugLevel)
	mainLog = log.WithFields(log.Fields{
		"name": "Main",
	})
}

func main() {
	pflag.Usage = func() {
		fmt.Printf(
			`Usage:    astron [options]... [CONFIG_FILE]
      
      Astron is a distributed server CLI.
      By default Astron looks for a configuration file in the current
      working directory as astrond.yml.  A different config file path
      can be specified as a positional argument.
      
      -h, --help      Print this help dialog.
      -v, --version   Print version information.
      -L, --log       Specify a file to write log messages to.
      -l, --loglevel  Specify the minimum log level that should be logged;
                        Error and Fatal levels will always be logged.
`)
		os.Exit(1)
	}

	logfilePtr := pflag.StringP("log", "L", "", "Specify the file to write log messages to.")
	loglevelPtr := pflag.StringP("loglevel", "l", "info", "Specify minimum log level that should be logged.")
	versionPtr := pflag.BoolP("version", "v", false, "Show the application version.")
	helpPtr := pflag.BoolP("help", "h", false, "Show the application usage.")

	pflag.Parse()

	if *helpPtr {
		pflag.Usage()
		os.Exit(1)
	}
	if *versionPtr {
		fmt.Printf(`
A Server Technology for Realtime Object Networking (Astron) in Golang
http://github.com/nosyliam/AstronGo

Revision: INDEV`)
		os.Exit(1)
	}
	if *loglevelPtr != "" {
		loglevelChoices := map[string]log.Level{"info": log.InfoLevel, "warning": log.WarnLevel, "error": log.ErrorLevel, "fatal": log.FatalLevel, "debug": log.DebugLevel}
		if choice, validChoice := loglevelChoices[*loglevelPtr]; !validChoice {
			mainLog.Fatal(fmt.Sprintf("Unknown log-level \"%s\".", *loglevelPtr))
			pflag.Usage()
			os.Exit(1)
		} else {
			log.SetLevel(choice)
		}
	}
	if *logfilePtr != "" {
		logfile, err := os.OpenFile(*logfilePtr, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			mainLog.Fatal(fmt.Sprintf("Failed to open log file \"%s\".", *logfilePtr))
			os.Exit(1)
		}
		logfile.Truncate(0)
		logfile.Seek(0, 0)

		defer logfile.Sync()
		defer logfile.Close()

		handler := core.NewMultiHandler(core.Log, core.NewLogger(logfile))
		log.SetHandler(handler)
	}

	var configPath, configName string
	args := pflag.Args()
	if len(args) > 0 {
		configName = filepath.Base(args[0])
		configName = strings.TrimSuffix(configName, path.Ext(configName))
		configPath = filepath.Dir(args[0])
	} else {
		configName = "astrond"
		configPath = "."
	}

	if err := core.LoadConfig(configPath, configName); err != nil {
		mainLog.Fatal(err.Error())
	}

	if err := core.LoadDC(); err != nil {
		mainLog.Fatal(err.Error())
	}

	hasher := dc.NewHashGenerator()
	core.DC.GenerateHash(hasher)
	core.Hash = hasher.Hash()
	mainLog.Info(fmt.Sprintf("DC hash: 0x%x", hasher.Hash()))
	eventlogger.StartEventLogger()
	messagedirector.Start()

	// Configure UberDOG list
	for _, ud := range core.Config.Uberdogs {
		class, ok := core.DC.ClassByName(ud.Class)
		if !ok {
			mainLog.Fatalf("For UberDOG %d, class %s does not exist!", ud.ID, ud.Class)
			return
		}

		core.Uberdogs = append(core.Uberdogs, core.Uberdog{
			Anonymous: ud.Anonymous,
			Id:        util.Doid_t(ud.ID),
			Class:     class,
		})
	}

	// Instantiate roles
	for _, role := range core.Config.Roles {
		switch role.Type {
		case "clientagent":
			clientagent.NewClientAgent(role)
		}
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	select {
	case sig := <-c:
		mainLog.Fatal(fmt.Sprintf("Got %s signal. Aborting...", sig))
		os.Exit(1)
	}
}
