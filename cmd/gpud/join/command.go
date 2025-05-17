// Package join implements the "join" command.
package join

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	apiv1 "github.com/leptonai/gpud/api/v1"
	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/asn"
	"github.com/leptonai/gpud/pkg/config"
	gpudstate "github.com/leptonai/gpud/pkg/gpud-state"
	"github.com/leptonai/gpud/pkg/log"
	pkgmachineinfo "github.com/leptonai/gpud/pkg/machine-info"
	latencyedge "github.com/leptonai/gpud/pkg/netutil/latency/edge"
	nvidianvml "github.com/leptonai/gpud/pkg/nvidia-query/nvml"
	"github.com/leptonai/gpud/pkg/sqlite"
)

// Command returns the cobra command for the "join" command.
func Command() *cobra.Command {
	return cmdRoot
}

var cmdRoot = &cobra.Command{
	Use:   "join",
	Short: "join gpud machine into a lepton cluster",
	RunE:  cmdRootFunc,
}

var (
	flagClusterName     string
	flagProvider        string
	flagNodeGroup       string
	flagExtraInfo       string
	flagGPUProduct      string
	flagRegion          string
	flagSkipInteractive bool
)

func init() {
	cmdRoot.PersistentFlags().StringVar(&flagClusterName, "cluster-name", "", "lepton.ai cluster name")
	cmdRoot.PersistentFlags().MarkDeprecated("cluster-name", "--cluster-name is deprecated")

	cmdRoot.PersistentFlags().StringVar(&flagProvider, "provider", "", "provider of the machine")
	cmdRoot.PersistentFlags().StringVar(&flagNodeGroup, "node-group", "", "node group of the machine")
	cmdRoot.PersistentFlags().StringVar(&flagExtraInfo, "extra-info", "", "extra info of the machine")
	cmdRoot.PersistentFlags().StringVar(&flagGPUProduct, "gpu-product", "unknown", "GPU shape of the machine")
	cmdRoot.PersistentFlags().StringVar(&flagRegion, "region", "unknown", "region of the machine")
	cmdRoot.PersistentFlags().BoolVar(&flagSkipInteractive, "skip-interactive", false, "skip interactive mode")
}

func cmdRootFunc(cmd *cobra.Command, args []string) error {
	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting join command")

	stateFile, err := config.DefaultStateFile()
	if err != nil {
		return fmt.Errorf("failed to get state file: %w", err)
	}

	dbRW, err := sqlite.Open(stateFile)
	if err != nil {
		return fmt.Errorf("failed to open state file: %w", err)
	}
	defer dbRW.Close()

	dbRO, err := sqlite.Open(stateFile, sqlite.WithReadOnly(true))
	if err != nil {
		return fmt.Errorf("failed to open state file: %w", err)
	}
	defer dbRO.Close()

	rootCtx, rootCancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer rootCancel()
	machineID, err := gpudstate.ReadMachineIDWithFallback(rootCtx, dbRW, dbRO)
	if err != nil {
		return err
	}

	// always read endpoint from state file
	endpoint, err := gpudstate.ReadMetadata(rootCtx, dbRO, gpudstate.MetadataKeyEndpoint)
	if err != nil {
		return fmt.Errorf("failed to read endpoint: %w", err)
	}
	if endpoint == "" {
		return errors.New("endpoint not found in state file")
	}

	// assume if not empty, it should have been persisted by the "gpud login" command
	privateIP, err := gpudstate.ReadMetadata(rootCtx, dbRO, gpudstate.MetadataKeyPrivateIP)
	if err != nil {
		return fmt.Errorf("failed to read private IP: %w", err)
	}

	// assume if not empty, it should have been persisted by the "gpud login" command
	publicIP, err := gpudstate.ReadMetadata(rootCtx, dbRO, gpudstate.MetadataKeyPublicIP)
	if err != nil {
		return fmt.Errorf("failed to read public IP: %w", err)
	}

	_, totalCPU, err := pkgmachineinfo.GetSystemResourceLogicalCores()
	if err != nil {
		return fmt.Errorf("failed to get system resource logical cores: %w", err)
	}

	nvmlInstance, err := nvidianvml.New()
	if err != nil {
		return err
	}
	if flagGPUProduct == "" {
		flagGPUProduct = nvmlInstance.ProductName()
	}

	// network section
	log.Logger.Debugw("measuring latencies to public tailscale DERP nodes to determine region")
	latencies, _ := latencyedge.Measure(rootCtx)
	if len(latencies) > 0 {
		closest := latencies.Closest()
		flagRegion = closest.RegionCode
	}

	detectProvider := "unknown"
	asnResult, err := asn.GetASLookup(publicIP)
	if err != nil {
		log.Logger.Errorf("failed to get asn lookup: %v", err)
	} else {
		detectProvider = asnResult.AsnName
	}

	if !flagSkipInteractive {
		reader := bufio.NewReader(os.Stdin)
		var input string
		if flagGPUProduct != "unknown" {
			fmt.Printf("We detect your gpu type is %v, if this is correct, press Enter. If not, please enter your gpu shape below\n", flagGPUProduct)
			input, err = reader.ReadString('\n')
			if err != nil {
				return err
			}
			if input != "\n" {
				flagGPUProduct = strings.TrimSpace(input)
			}
		}

		fmt.Printf("We detect your public IP is %v, if this is correct, press Enter. If not, please enter your public IP below\n", publicIP)
		input, err = reader.ReadString('\n')
		if err != nil {
			return err
		}
		if input != "\n" {
			publicIP = strings.TrimSpace(input)
		}

		if flagProvider == "" {
			fmt.Printf("Provider name not specified, we detected your provider is %v, if correct, press Enter. If not, please enter your provider's name below\n", detectProvider)
			input, err = reader.ReadString('\n')
			if err != nil {
				return err
			}
			if input != "\n" {
				flagProvider = strings.TrimSpace(input)
			} else {
				flagProvider = detectProvider
			}
		}

		fmt.Printf("We detect your region is %v, if this is correct, press Enter. If not, please enter your region below\n", flagRegion)
		input, err = reader.ReadString('\n')
		if err != nil {
			return err
		}
		if input != "\n" {
			flagRegion = strings.TrimSpace(input)
		}
	} else {
		if flagProvider == "" {
			flagProvider = detectProvider
		}
	}

	fmt.Printf("%sWarning: GPUd will upgrade your container runtime to containerd, will affect your current running containers (if any)%s\n", "\033[33m", "\033[0m")
	fmt.Printf("%sWarning: GPUd will Reboot your machine to finish necessary setup%s\n", "\033[33m", "\033[0m")
	fmt.Printf("Please look carefully about the above warning, if ok, please hit Enter\n")

	content := apiv1.JoinRequest{
		ID:               machineID,
		ClusterName:      flagClusterName,
		PublicIP:         publicIP,
		Provider:         strings.Replace(flagProvider, " ", "-", -1),
		ProviderGPUShape: flagGPUProduct,
		TotalCPU:         totalCPU,
		NodeGroup:        flagNodeGroup,
		ExtraInfo:        flagExtraInfo,
		Region:           flagRegion,
		PrivateIP:        privateIP,
	}

	rawPayload, _ := json.Marshal(&content)
	fmt.Println("Your machine will be initialized with following configuration, please press Enter if it is ok")
	prettyJSON, _ := json.MarshalIndent(content, "", "  ")
	fmt.Println(string(prettyJSON))

	if !flagSkipInteractive {
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		if input != "\n" {
			fmt.Println("Non empty input received, GPUd join aborted.")
			return nil
		}
	}

	fmt.Println("Please wait while control plane is initializing basic setup for your machine, this may take up to one minute...")
	response, err := http.Post(createJoinURL(endpoint), "application/json", bytes.NewBuffer(rawPayload))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}
		var errorResponse apiv1.JoinResponse
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			return fmt.Errorf("error parsing error response: %v %s", err, string(body))
		}
		return fmt.Errorf("failed to join: %v", errorResponse)
	}

	// persist on the successful join
	// so that next gpud up/run doesn't need to specify the same parameters
	if err := gpudstate.SetMetadata(rootCtx, dbRW, gpudstate.MetadataKeyPublicIP, publicIP); err != nil {
		return fmt.Errorf("failed to record public IP: %w", err)
	}
	if err := gpudstate.SetMetadata(rootCtx, dbRW, gpudstate.MetadataKeyProvider, flagProvider); err != nil {
		return fmt.Errorf("failed to record provider: %w", err)
	}
	if err := gpudstate.SetMetadata(rootCtx, dbRW, gpudstate.MetadataKeyNodeGroup, flagNodeGroup); err != nil {
		return fmt.Errorf("failed to record node group: %w", err)
	}
	if err := gpudstate.SetMetadata(rootCtx, dbRW, gpudstate.MetadataKeyRegion, flagRegion); err != nil {
		return fmt.Errorf("failed to record region: %w", err)
	}
	if err := gpudstate.SetMetadata(rootCtx, dbRW, gpudstate.MetadataKeyExtraInfo, flagExtraInfo); err != nil {
		return fmt.Errorf("failed to record extra info: %w", err)
	}

	fmt.Println("Basic setup finished, GPUd is installing necessary components onto your machine, this may take 10 - 15 minutes.\nYou can run `gpud status` or `gpud status -w` to check the progress of each component.")
	return nil
}

// createJoinURL creates a URL for the join endpoint
func createJoinURL(endpoint string) string {
	host := endpoint
	url, _ := url.Parse(endpoint)
	if url.Host != "" {
		host = url.Host
	}
	return fmt.Sprintf("https://%s/api/v1/join", host)
}
