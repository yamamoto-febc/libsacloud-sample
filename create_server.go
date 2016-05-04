package main

import (
	"fmt"
	API "github.com/yamamoto-febc/libsacloud/api"
	"os"
	"time"
)

func main() {

	// settings
	var (
		token        = os.Args[1]
		secret       = os.Args[2]
		zone         = os.Args[3]
		name         = "libsacloud demo"
		description  = "libsacloud demo description"
		tag          = "libsacloud-test"
		cpu          = 1
		mem          = 2
		hostName     = "libsacloud-test"
		password     = "C8#mf92mp!*s"
		sshPublicKey = "ssh-rsa AAAA..."
	)

	// authorize
	api := API.NewClient(token, secret, zone)

	//search archives
	fmt.Println("searching archives")
	res, _ := api.Archive.
		WithNameLike("CentOS 6.7 64bit").
		WithSharedScope().
		Limit(1).
		Find()

	archive := res.Archives[0]

	// search scripts
	fmt.Println("searching scripts")
	res, _ = api.Note.
		WithNameLike("WordPress").
		WithSharedScope().
		Limit(1).
		Find()
	script := res.Notes[0]

	// create a disk
	fmt.Println("creating a disk")
	disk := api.Disk.New()
	disk.Name = name
	disk.Name = name
	disk.Description = description
	disk.Tags = []string{tag}
	disk.SetDiskPlanToSSD()
	disk.SetSourceArchive(archive.ID)

	disk, _ = api.Disk.Create(disk)

	// create a server
	fmt.Println("creating a server")
	server := api.Server.New()
	server.Name = name
	server.Description = description
	server.Tags = []string{tag}

	// (set ServerPlan)
	plan, _ := api.Product.Server.GetBySpec(cpu, mem)
	server.SetServerPlanByID(plan.ID.String())

	server, _ = api.Server.Create(server)

	// connect to shared segment

	fmt.Println("connecting the server to shared segment")
	iface, _ := api.Interface.CreateAndConnectToServer(server.ID)
	api.Interface.ConnectToSharedSegment(iface.ID)

	// wait disk copy
	err := api.Disk.SleepWhileCopying(disk.ID, 120*time.Second)
	if err != nil {
		fmt.Println("failed")
		os.Exit(1)
	}

	// config the disk
	diskconf := api.Disk.NewCondig()
	diskconf.HostName = hostName
	diskconf.Password = password
	diskconf.SSHKey.PublicKey = sshPublicKey
	diskconf.AddNote(script.ID)
	api.Disk.Config(disk.ID, diskconf)

	// boot
	fmt.Println("booting the server")
	api.Server.Boot(server.ID)

	// stop
	time.Sleep(3 * time.Second)
	fmt.Println("stopping the server")
	api.Server.Stop(server.ID)

	err = api.Server.SleepUntilDown(server.ID, 120*time.Second)
	if err != nil {
		fmt.Println("failed")
		os.Exit(1)
	}

	// disconnect the disk from the server
	fmt.Println("disconnecting the disk")
	api.Disk.DisconnectFromServer(disk.ID)

	// delete the server
	fmt.Println("deleting the server")
	api.Server.Delete(server.ID)

	// delete the disk
	fmt.Println("deleting the disk")
	api.Disk.Delete(disk.ID)

}
