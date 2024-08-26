package services

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// install
func Import(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	if len(args) != 2 {
		fmt.Println("Please give me: eris services install [name] [location]")
		return
	}
	IfExit(ImportServiceRaw(args[0], args[1]))
}

func New(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	if len(args) != 2 {
		fmt.Println("Please give me: eris new [name] [containerImage]")
		return
	}
	IfExit(NewServiceRaw(args[0], args[1]))
}

func Edit(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	EditServiceRaw(args[0])
}

func Rename(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	if len(args) != 2 {
		fmt.Println("Please give me: eris services rename [oldName] [newName]")
		return
	}
	IfExit(RenameServiceRaw(args[0], args[1]))
}

func Inspect(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	if len(args) == 1 {
		args = append(args, "all")
	}
	IfExit(InspectServiceRaw(args[0], args[1]))
}

func Export(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	IfExit(ExportServiceRaw(args[0]))
}

// Updates an installed service, or installs it if it has not been installed.
func Update(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	IfExit(UpdateServiceRaw(args[0]))
}

// list known
func ListKnown() {
	services := ListKnownRaw()
	for _, s := range services {
		fmt.Println(s)
	}
}

func ListRunning() {
	services := ListRunningRaw()
	for _, s := range services {
		fmt.Println(s)
	}
}

func ListExisting() {
	services := ListExistingRaw()
	for _, s := range services {
		fmt.Println(s)
	}
}

func Rm(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	IfExit(RmServiceRaw(args[0], cmd.Flags().Lookup("force").Changed))
}

func ImportServiceRaw(servName, servPath string) error {
	fileName := filepath.Join(ServicesPath, servName)
	if filepath.Ext(fileName) == "" {
		fileName = fileName + ".toml"
	}

	s := strings.Split(servPath, ":")
	if s[0] == "ipfs" {

		var err error
		if logger.Level > 0 {
			err = util.GetFromIPFS(s[1], fileName, logger.Writer)
		} else {
			err = util.GetFromIPFS(s[1], fileName, bytes.NewBuffer([]byte{}))
		}

		if err != nil {
			return err
		}
		return nil
	}

	if strings.Contains(s[0], "github") {
		logger.Errorln("https://twitter.com/ryaneshea/status/595957712040628224")
		return nil
	}

	return fmt.Errorf("I do not know how to get that file. Sorry.")
}

func NewServiceRaw(servName, imageName string) error {
	srv := &def.Service{
		Name:  servName,
		Image: imageName,
	}
	srvDef := &def.ServiceDefinition{
		Service:    srv,
		Maintainer: &def.Maintainer{},
		Location:   &def.Location{},
		Machine:    &def.Machine{},
	}

	err := WriteServiceDefinitionFile(srvDef, "")
	if err != nil {
		return err
	}
	return nil
}

func EditServiceRaw(servName string) {
	Editor(servDefFileByServName(servName))
}

func RenameServiceRaw(oldName, newName string) error {
	if parseKnown(oldName) {
		logger.Infoln("Renaming service", oldName, "to", newName)

		serviceDef, err := LoadServiceDefinition(oldName)
		if err != nil {
			return err
		}
		err = perform.DockerRename(serviceDef.Service, serviceDef.Operations, oldName, newName)
		if err != nil {
			return err
		}
		oldFile := servDefFileByServName(oldName)
		newFile := strings.Replace(oldFile, oldName, newName, 1)

		serviceDef.Service.Name = newName
		err = WriteServiceDefinitionFile(serviceDef, newFile)
		if err != nil {
			return err
		}

		err = data.RenameDataRaw(oldName, newName)
		if err != nil {
			return err
		}

		os.Remove(oldFile)
	} else {
		return fmt.Errorf("I cannot find that service. Please check the service name you sent me.")
	}

	return nil
}

func InspectServiceRaw(servName, field string) error {
	service, err := LoadServiceDefinition(servName)
	if err != nil {
		return err
	}
	err = InspectServiceByService(service.Service, service.Operations, field)
	if err != nil {
		return err
	}
	return nil
}

func ExportServiceRaw(servName string) error {
	if parseKnown(servName) {
		ipfsService, err := LoadServiceDefinition("ipfs")
		if err != nil {
			return err
		}

		if IsServiceRunning(ipfsService.Service) {
			logger.Infoln("IPFS is running. Adding now.")

			hash, err := exportFile(servName)
			if err != nil {
				return err
			}

			logger.Errorln(hash)
		} else {
			logger.Infoln("IPFS is not running. Starting now.")
			err := StartServiceByService(ipfsService.Service, ipfsService.Operations)
			if err != nil {
				return err
			}

			hash, err := exportFile(servName)
			if err != nil {
				return err
			}

			logger.Errorln(hash)
		}

	} else {
		return fmt.Errorf(`I don't known of that service.
Please retry with a known service.
To find known services use: eris services known`)
	}
	return nil
}

func InspectServiceByService(srv *def.Service, ops *def.ServiceOperation, field string) error {
	if IsServiceExisting(srv) {
		err := perform.DockerInspect(srv, ops, field)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("No service matching that name.")
	}
	return nil
}

func ListKnownRaw() []string {
	srvs := []string{}
	fileTypes := []string{}
	for _, t := range []string{"*.json", "*.yaml", "*.toml"} {
		fileTypes = append(fileTypes, filepath.Join(ServicesPath, t))
	}
	for _, t := range fileTypes {
		s, _ := filepath.Glob(t)
		for _, s1 := range s {
			s1 = strings.Split(filepath.Base(s1), ".")[0]
			srvs = append(srvs, s1)
		}
	}
	return srvs
}

func ListRunningRaw() []string {
	return listServices(false)
}

func ListExistingRaw() []string {
	return listServices(true)
}

func UpdateServiceRaw(servName string) error {
	service, err := LoadServiceDefinition(servName)
	if err != nil {
		return err
	}
	err = perform.DockerRebuild(service.Service, service.Operations, true)
	if err != nil {
		return err
	}
	return nil
}

func RmServiceRaw(servName string, force bool) error {
	service, err := LoadServiceDefinition(servName)
	if err != nil {
		return err
	}
	err = perform.DockerRemove(service.Service, service.Operations)
	if err != nil {
		return err
	}
	if force {
		oldFile := servDefFileByServName(servName)
		os.Remove(oldFile)
	}
	return nil
}

func exportFile(servName string) (string, error) {
	fileName := servDefFileByServName(servName)

	var err error
	var hash string
	if logger.Level > 0 {
		hash, err = util.SendToIPFS(fileName, logger.Writer)
	} else {
		hash, err = util.SendToIPFS(fileName, bytes.NewBuffer([]byte{}))
	}

	if err != nil {
		return "", err
	}

	return hash, nil
}
