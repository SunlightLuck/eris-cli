package services

import (
	"fmt"
	"os"
	"path"
	"strings"
	"strconv"
	"testing"

	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/log"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/util"
)

var srv *def.ServiceDefinition
var erisDir string = path.Join(os.TempDir(), "eris")
var servName string = "ipfs"
var hash string

func TestMain(m *testing.M) {
	var logLevel int

	if os.Getenv("LOG_LEVEL") != "" {
		logLevel, _ = strconv.Atoi(os.Getenv("LOG_LEVEL"))
	} else {
		logLevel = 0
		// logLevel = 1
		// logLevel = 2
	}
	log.SetLoggers(logLevel, os.Stdout, os.Stderr)

	ifExit(testsInit())

	exitCode := m.Run()

	logger.Infoln("Commensing with Tests Tear Down.")
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		ifExit(testsTearDown())
	}

	os.Exit(exitCode)
}

func TestKnownServiceRaw(t *testing.T) {
	do := def.NowDo()
	ifExit(ListKnownRaw(do))
	k := strings.Split(do.Result, "\n") // tests output formatting.

	if len(k) != 2 {
		ifExit(fmt.Errorf("More than two service definitions found. Something is wrong.\n"))
	}

	if k[0] != "ipfs" {
		ifExit(fmt.Errorf("Could not find ipfs service definition. Services found =>\t%v\n", k))
	}
}

func TestLoadServiceDefinition(t *testing.T) {
	var e error
	srv, e = loaders.LoadServiceDefinition(servName, 1)
	if e != nil {
		logger.Errorln(e)
		t.FailNow()
	}

	if srv.Name != servName {
		logger.Errorf("FAILURE: improper name on LOAD. expected: %s\tgot: %s\n", servName, srv.Name)
	}

	if srv.Service.Name != servName {
		logger.Errorf("FAILURE: improper service name on LOAD. expected: %s\tgot: %s\n", servName, srv.Service.Name)
		t.FailNow()
	}

	if !srv.Service.AutoData {
		logger.Errorf("FAILURE: data_container not properly read on LOAD.\n")
		t.FailNow()
	}

	if srv.Operations.DataContainerName == "" {
		logger.Errorf("FAILURE: data_container_name not set.\n")
		t.Fail()
	}
}

func TestStartServiceRaw(t *testing.T) {
	do := def.NowDo()
	do.Args = []string{servName}
	do.Operations.ContainerNumber = 1
	logger.Debugf("Starting service (via tests) =>\t%s\n", servName)
	e := StartServiceRaw(do)
	if e != nil {
		logger.Infoln("Error starting service =>\t%v\n", e)
		t.Fail()
	}

	testExistAndRun(t, servName, 1, true, true)
}

func TestInspectServiceRaw(t *testing.T) {
	do := def.NowDo()
	do.Name = servName
	do.Args = []string{"name"}
	do.Operations.ContainerNumber = 1
	logger.Debugf("Inspect service (via tests) =>\t%s:%v\n", servName, do.Args)
	e := InspectServiceRaw(do)
	if e != nil {
		logger.Infof("Error inspecting service =>\t%v\n", e)
		t.FailNow()
	}

	do = def.NowDo()
	do.Name = servName
	do.Args = []string{"config.user"}
	do.Operations.ContainerNumber = 1
	logger.Debugf("Inspect service (via tests) =>\t%s:%v\n", servName, do.Args)
	e = InspectServiceRaw(do)
	if e != nil {
		logger.Infof("Error inspecting service =>\t%v\n", e)
		t.Fail()
	}
}

func TestLogsServiceRaw(t *testing.T) {
	do := def.NowDo()
	do.Name = servName
	do.Follow = false
	do.Tail = "all"
	logger.Debugf("Inspect logs (via tests) =>\t%s:%v\n", servName, do.Tail)
	e := LogsServiceRaw(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}
}

func TestExecServiceRaw(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have exec privileges (due to their driver). Skipping test.")
		return
	}

	do := def.NowDo()
	do.Name = servName
	do.Interactive = false
	do.Args = strings.Fields("ls -la /root/")
	logger.Debugf("Exec-ing serv (via tests) =>\t%s:%v\n", servName, strings.Join(do.Args, " "))
	e := ExecServiceRaw(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}
}

func TestUpdateServiceRaw(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	do := def.NowDo()
	do.Name = servName
	do.SkipPull = true
	logger.Debugf("Update serv (via tests) =>\t%s\n", servName)
	e := UpdateServiceRaw(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testExistAndRun(t, servName, 1, true, true)
}

func TestKillServiceRaw(t *testing.T) {
	do := def.NowDo()
	do.Name = servName
	do.Rm = false
	do.RmD = false
	do.Args = []string{servName}
	logger.Debugf("Stopping serv (via tests) =>\t%s\n", servName)
	e := KillServiceRaw(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testExistAndRun(t, servName, 1, true, false)
}

func TestRmServiceRaw(t *testing.T) {
	if os.Getenv("TEST_IN_CIRCLE") == "true" {
		logger.Println("Testing in Circle. Where we don't have rm privileges (due to their driver). Skipping test.")
		return
	}

	do := def.NowDo()
	do.Name = servName
	do.Args = []string{servName}
	do.File = false
	do.RmD = true
	logger.Debugf("Removing serv (via tests) =>\t%s\n", servName)
	e := RmServiceRaw(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testExistAndRun(t, servName, 1, false, false)
}

func TestNewServiceRaw(t *testing.T) {
	do := def.NowDo()
	do.Name = "keys"
	do.Args = []string{"eris/keys"}
	logger.Debugf("New-ing serv (via tests) =>\t%s:%v\n", do.Name, do.Args)
	e := NewServiceRaw(do)
	if e != nil {
		logger.Errorln(e)
		t.FailNow()
	}

	do = def.NowDo()
	do.Args = []string{"keys"}
	do.Operations.ContainerNumber = 1
	logger.Debugf("Stating serv (via tests) =>\t%v:%d\n", do.Args, do.Operations.ContainerNumber)
	e = StartServiceRaw(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testExistAndRun(t, "keys", 1, true, true)
}

func TestRenameServiceRaw(t *testing.T) {
	do := def.NowDo()
	do.Name = "keys"
	do.NewName = "syek"
	do.Operations.ContainerNumber = 1
	logger.Debugf("Renaming serv (via tests) =>\t%s:%v\n", do.Name, do.NewName)
	e := RenameServiceRaw(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testExistAndRun(t, "syek", 1, true, true)

	do = def.NowDo()
	do.Name = "syek"
	do.NewName = "keys"
	do.Operations.ContainerNumber = 1
	logger.Debugf("Renaming serv (via tests) =>\t%s:%v\n", do.Name, do.NewName)
	e = RenameServiceRaw(do)
	if e != nil {
		logger.Errorln(e)
		t.Fail()
	}

	testExistAndRun(t, "keys", 1, true, true)
}

// tests remove+kill
func TestKillServiceRawPostNew(t *testing.T) {

	do := def.NowDo()
	do.Args = []string{"keys"}
	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		do.Rm = true
		do.RmD = true
	}
	logger.Debugf("Renaming serv (via tests) =>\t%s:%v\n", do.Name, do.NewName)
	e := KillServiceRaw(do)
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}

	if os.Getenv("TEST_IN_CIRCLE") != "true" {
		testExistAndRun(t, "keys", 1, false, false)
	} else {
		testExistAndRun(t, "keys", 1, true, false)
	}
}

func testExistAndRun(t *testing.T, servName string, containerNumber int, toExist, toRun bool) {
	var exist, run bool
	logger.Infof("\nTesting whether (%s) is running? (%t) and existing? (%t)\n", servName, toRun, toExist)
	servName = util.ServiceContainersName(servName, containerNumber)

	do := def.NowDo()
	do.Quiet = true
	do.Args = []string{"testing"}
	if err := ListExistingRaw(do); err != nil {
		logger.Errorln(err)
		t.FailNow()
	}
	res := strings.Split(do.Result, "\n")
	for _, r := range res {
		logger.Debugf("Existing =>\t\t\t%s\n", r)
		if r == util.ContainersShortName(servName) {
			exist = true
		}
	}

	do = def.NowDo()
	do.Quiet = true
	do.Args = []string{"testing"}
	if err := ListRunningRaw(do); err != nil {
		logger.Errorln(err)
		t.FailNow()
	}
	res = strings.Split(do.Result, "\n")
	for _, r := range res {
		logger.Debugf("Running =>\t\t\t%s\n", r)
		if r == util.ContainersShortName(servName) {
			run = true
		}
	}

	if toExist != exist {
		if toExist {
			logger.Infof("Could not find an existing =>\t%s\n", servName)
		} else {
			logger.Infof("Found an existing instance of %s when I shouldn't have\n", servName)
		}
		t.Fail()
	}

	if toRun != run {
		if toRun {
			logger.Infof("Could not find a running =>\t%s\n", servName)
		} else {
			logger.Infof("Found a running instance of %s when I shouldn't have\n", servName)
		}
		t.Fail()
	}

	logger.Debugln("")
}

func testsInit() error {
	var err error
	// TODO: make a reader/pipe so we can see what is written from tests.
	util.GlobalConfig, err = util.SetGlobalObject(os.Stdout, os.Stderr)
	ifExit(err)

	// common is initialized on import so
	// we have to manually override these
	// variables to ensure that the tests
	// run correctly.
	util.ChangeErisDir(erisDir)

	// this dumps the ipfs service def into the temp dir which
	// has been set as the erisRoot
	ifExit(util.Initialize(false, false))

	// init dockerClient
	util.DockerConnect(false)

	// set ipfs endpoint
	os.Setenv("ERIS_IPFS_HOST", "http://0.0.0.0")

	// make sure ipfs not running
	do := def.NowDo()
	do.Quiet = true
	logger.Debugln("Finding the running services.")
	if err := ListRunningRaw(do); err != nil {
		ifExit(err)
	}
	res := strings.Split(do.Result, "\n")
	for _, r := range res {
		if r == "ipfs" {
			ifExit(fmt.Errorf("IPFS service is running.\nPlease stop it with.\neris services stop -rx ipfs\n"))
		}
	}
	// make sure ipfs container does not exist
	do = def.NowDo()
	do.Quiet = true
	if err := ListExistingRaw(do); err != nil {
		ifExit(err)
	}
	res = strings.Split(do.Result, "\n")
	for _, r := range res {
		if r == "ipfs" {
			ifExit(fmt.Errorf("IPFS service exists.\nPlease remove it with\neris services rm ipfs\n"))
		}
	}

	logger.Infoln("Test init completed. Starting main test sequence now.")
	return nil
}

func testsTearDown() error {
	return os.RemoveAll(erisDir)
}

func ifExit(err error) {
	if err != nil {
		logger.Errorln(err)
		log.Flush()
		testsTearDown()
		os.Exit(1)
	}
}