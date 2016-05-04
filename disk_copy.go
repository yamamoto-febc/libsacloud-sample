package main

import (
	"fmt"
	"github.com/webguerilla/ftps"
	API "github.com/yamamoto-febc/libsacloud/api"
	"os"
	"time"
)

func main() {

	// settings
	var (
		token   = os.Args[1]
		secret  = os.Args[2]
		zone    = os.Args[3]
		srcName = "GitLab"
	)

	// authorize
	api := API.NewClient(token, secret, zone)

	// search the source disk
	res, _ := api.Disk.
		WithNameLike(srcName).
		Limit(1).
		Find()
	if res.Count == 0 {
		panic("Disk `GitLab` not found")
	}

	disk := res.Disks[0]

	// copy the disk to a new archive
	fmt.Println("copying the disk to a new archive")

	archive := api.Archive.New()
	archive.Name = fmt.Sprintf("Copy:%s", disk.Name)
	archive.SetSourceDisk(disk.ID)
	archive, _ = api.Archive.Create(archive)
	api.Archive.SleepWhileCopying(archive.ID, 180*time.Second)

	// get FTP information
	ftp, _ := api.Archive.OpenFTP(archive.ID, false)
	fmt.Println("FTP information:")
	fmt.Println("  user: " + ftp.User)
	fmt.Println("  pass: " + ftp.Password)
	fmt.Println("  host: " + ftp.HostName)

	// download the archive via FTPS
	ftpsClient := &ftps.FTPS{}
	ftpsClient.TLSConfig.InsecureSkipVerify = true
	ftpsClient.Connect(ftp.HostName, 21)
	ftpsClient.Login(ftp.User, ftp.Password)
	err := ftpsClient.RetrieveFile("archive.img", "archive.img")
	if err != nil {
		panic(err)
	}
	ftpsClient.Quit()

	// delete the archive after download
	fmt.Println("deleting the archive")
	api.Archive.CloseFTP(archive.ID)
	api.Archive.Delete(archive.ID)

}
