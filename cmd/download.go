package cmd

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"net"
	"strings"
	"time"
)

// Slice of strings with placeholder text.
var fakeInstallList = strings.Split("pseudo-excel pseudo-photoshop pseudo-chrome pseudo-outlook pseudo-explorer "+
	"pseudo-dops pseudo-git pseudo-vsc pseudo-intellij pseudo-minecraft pseudo-scoop pseudo-chocolatey", " ")

func init() {
	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = &cobra.Command{
	Use:     "download",
	Aliases: []string{"dl"},
	Short:   "Download a mod",
	RunE: func(cmd *cobra.Command, args []string) error {

		conn, err := net.Dial("udp", "127.0.0.1:15777")
		if err != nil {
			return err
		}
		defer conn.Close()

		const protoVersion = 0
		encoded := make([]byte, 8)
		binary.LittleEndian.PutUint64(encoded, uint64(time.Now().UnixMilli()))

		query := append([]byte{0, protoVersion}, encoded...)
		spew.Dump("Query:", query)

		if _, err := conn.Write(query); err != nil {
			return err
		}

		response := make([]byte, 17)
		_, err = bufio.NewReader(conn).Read(response)

		spew.Dump("Response:", response)

		serverQueryID := response[0]
		serverProtocolVersion := response[1]
		serverTimestamp := binary.LittleEndian.Uint64(response[2:10])
		serverState := response[10]
		serverNetCL := binary.LittleEndian.Uint32(response[11:15])
		beaconPort := binary.LittleEndian.Uint16(response[15:])

		fmt.Printf("Server query ID: %d\n", serverQueryID)
		fmt.Printf("Server protocol version: %d\n", serverProtocolVersion)
		fmt.Printf("Server timestamp: %d\n", serverTimestamp)
		fmt.Printf("Server state: %d\n", serverState)
		fmt.Printf("Server net CL: %d\n", serverNetCL)
		fmt.Printf("Server beacon port: %d\n", beaconPort)
		//for i := 0; i < 5; i++ {
		//	log.Info().Int("i", i).Msg("Foo")
		//	time.Sleep(time.Second)
		//}
		//
		//p, _ := pterm.DefaultProgressbar.WithTotal(len(fakeInstallList)).WithTitle("Downloading stuff").Start()
		//
		//for i := 0; i < p.Total; i++ {
		//	p.UpdateTitle("Downloading " + fakeInstallList[i])         // Update the title of the progressbar.
		//	pterm.Success.Println("Downloading " + fakeInstallList[i]) // If a progressbar is running, each print will be printed above the progressbar.
		//	p.Increment()                                              // Increment the progressbar by one. Use Add(x int) to increment by a custom amount.
		//	time.Sleep(time.Millisecond * 350)                         // Sleep 350 milliseconds.
		//}
		//
		//for i := 0; i < 5; i++ {
		//	log.Info().Int("i", i).Msg("Bar")
		//	time.Sleep(time.Second)
		//}

		return nil
	},
}
