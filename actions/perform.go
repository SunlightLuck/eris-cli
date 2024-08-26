package actions

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/eris-ltd/eris-cli/chains"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/services"
)

func DoRaw(do *definitions.Do) error {
	var err error
	var actionVars []string
	do.Action, actionVars, err = LoadActionDefinition(strings.Join(do.Args, "_"))
	if err != nil {
		return err
	}

	if err := MergeStepsAndCLIArgs(do.Action, &actionVars, do.Args); err != nil {
		return err
	}

	if err := StartServicesAndChains(do); err != nil {
		return err
	}

	if err := PerformCommand(do.Action, actionVars, do.Quiet); err != nil {
		return err
	}

	return nil
}

func StartServicesAndChains(do *definitions.Do) error {
	// start the services and chains
	doSrvs := definitions.NowDo()
	doSrvs.Args = do.Action.ServiceDeps
	logger.Debugf("Starting Services. Args =>\t%v\n", doSrvs.Args)
	if err := services.StartServiceRaw(doSrvs); err != nil {
		return err
	}

	doChns := definitions.NowDo()
	doChns.Name = do.ChainName
	logger.Debugf("Starting Chain. Name =>\t%v\n", doChns.Name)
	if err := chains.StartChainRaw(do); err != nil {
		return err
	}

	return nil
}

func PerformCommand(action *definitions.Action, actionVars []string, quiet bool) error {
	logger.Infof("Performing Action =>\t\t%s.\n", action.Name)

	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	logger.Debugf("Directory for action =>\t\t%s\n", dir)

	// pull actionVars (first given from command line) and
	// combine with the environment variables (given in the
	// action definition files) and finally combine with
	// the hosts os.Environ() to provide the full set of
	// variables to be consumed during the steps phase.
	for k, v := range action.Environment {
		actionVars = append(actionVars, fmt.Sprintf("%s=%s", k, v))
	}

	for _, v := range actionVars {
		logger.Debugf("Variable for action =>\t\t%s\n", v)
	}

	actionVars = append(os.Environ(), actionVars...)

	for n, step := range action.Steps {
		cmd := exec.Command("sh", "-c", step)
		cmd.Env = actionVars
		cmd.Dir = dir

		logger.Debugf("Performing Step %d =>\t\t%s\n", n+1, strings.Join(cmd.Args, " "))

		prev, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("error running command (%v): %s", err, prev)
		}

		if !quiet {
			logger.Println(strings.TrimSpace(string(prev)))
		}

		if n != 0 {
			actionVars = actionVars[:len(actionVars)-1]
		}
		actionVars = append(actionVars, ("prev=" + strings.TrimSpace(string(prev))))
	}

	logger.Infoln("Action performed")
	return nil
}
