package utils

import (
  "os"
  "fmt"
  "io"
  "strings"
  "errors"
  "path/filepath"
)

func RemoveFile(filePath string, logger Logger) (error){
  err := os.Remove(filePath)
  if err != nil{
      logger.Log(fmt.Sprintf("Error deleting file %s: %v", filePath, err))
      return err
  }
  return nil
}

func MoveFile(fileFullPath string, baseDir string, logger Logger) (error){
  if !strings.HasPrefix(fileFullPath, "/home/"){
    return errors.New("Cannot move file, invalid origin")
  }

  dir := filepath.Dir(fileFullPath)
  quarantineFullPath := fmt.Sprintf("%s/%s", baseDir, dir)
  newName := fmt.Sprintf("%s%s", baseDir, fileFullPath)

  err := os.MkdirAll(quarantineFullPath, 0755)
  if err != nil{
      logger.Log(fmt.Sprintf("Error creating Quaratines full path %s: %v", quarantineFullPath, err))
      return err
  }

  sf, err := os.Open(fileFullPath)
  if err != nil {
      logger.Log(fmt.Sprintf("Error reading file to move to Quarantine %s: %v", fileFullPath, err))
      return err
  }
  defer sf.Close()

  df, err := os.OpenFile(newName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0000)
  if err != nil {
    logger.Log(fmt.Sprintf("Error creating file on Quarantine %s: %v", newName, err))
    return err
  }
  defer df.Close()

  _, err = io.Copy(df, sf)
  if err != nil{
    logger.Log(fmt.Sprintf("Error moving file to Quarantine %s: %v", fileFullPath, err))
    return err
  } else {
    return RemoveFile(fileFullPath, logger)
  }

}

