package services

import (
  "fmt"
  "os"
  "strings"
  "strconv"

  def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"
  dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/viper"
)

func LoadServiceDefinition(servName string) (*def.ServiceDefinition) {
  var service def.ServiceDefinition
  serviceConf := loadServiceDefinition(servName)

  // marshal service and always reset the operational requirements
  // this will make sure to sync with docker so that if changes
  // have occured in the interim they are caught.
  marshalServiceDefinition(serviceConf, &service)
  service.Operations = &def.ServiceOperation{}

  checkServiceHasImage(service.Service)
  checkServiceHasName(service.Service, service.Operations)
  checkServiceHasDataContainer(serviceConf, service.Service, service.Operations)
  checkDataContainerHasName(service.Operations)

  return &service
}

func LoadService(servName string) (*def.Service) {
  sd := LoadServiceDefinition(servName)
  return sd.Service
}

func LoadServiceOperation(servName string) (*def.ServiceOperation) {
  sd := LoadServiceDefinition(servName)
  return sd.Operations
}

func loadServiceDefinition(servName string) *viper.Viper {
  var serviceConf = viper.New()

  serviceConf.AddConfigPath(dir.ServicesPath)
  serviceConf.SetConfigName(servName)
  serviceConf.ReadInConfig()

  return serviceConf
}

func servDefFileByServName(servName string) string {
  serviceConf := loadServiceDefinition(servName)
  return serviceConf.ConfigFileUsed()
}

func marshalServiceDefinition(serviceConf *viper.Viper, service *def.ServiceDefinition) {
  err := serviceConf.Marshal(service)
  if err != nil {
    // TODO: error handling
    fmt.Println(err)
    os.Exit(1)
  }
}

func checkServiceGiven(args []string) {
  if len(args) == 0 {
    // TODO: betterly error handling
    fmt.Println("No Service Given. Please rerun command with a known service.")
    os.Exit(1)
  }
}

func checkServiceHasImage(service *def.Service) {
  // Services must be given an image. Flame out if they do not.
  if service.Image == "" {
    fmt.Println("An \"image\" field is required in the service definition file.")
    os.Exit(1)
  }
}

func checkServiceHasName(service *def.Service, ops *def.ServiceOperation) {
  // If no name use image name
  if service.Name == "" {
    if service.Image != "" {
      service.Name = strings.Replace(service.Image, "/", "_", -1)
    }
  }

  containerNumber := 1 // tmp
  ops.SrvContainerName = "eris_service_" + service.Name + "_" + strconv.Itoa(containerNumber)
}

func checkServiceHasDataContainer(serviceConf *viper.Viper, service *def.Service, ops *def.ServiceOperation) {
  // toml bools don't really marshal well
  if serviceConf.GetBool("service.data_container") {
    service.AutoData = true
    ops.DataContainer = true
  }
}

func checkDataContainerHasName(ops *def.ServiceOperation) {
  if ops.DataContainer {
    ops.DataContainerName = ""
    if ops.DataContainer {
      dataSplit := strings.Split(ops.SrvContainerName, "_")
      dataSplit[1] = "data"
      ops.DataContainerName = strings.Join(dataSplit, "_")
    }
  }
}