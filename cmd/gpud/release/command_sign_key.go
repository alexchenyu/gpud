package release

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/log"
	"github.com/leptonai/gpud/pkg/release/distsign"
)

var cmdSignKey = &cobra.Command{
	Use:   "sign-key",
	Short: "sign signing keys with a root key",
	RunE:  cmdSignKeyFunc,
}

var (
	flagSignKeyRootPrivPath string
	flagSignKeySignPubPath  string
	flagSignKeySigPath      string
)

func init() {
	cmdSignKey.PersistentFlags().StringVar(&flagSignKeyRootPrivPath, "root-priv-path", "", "path of root private key")
	cmdSignKey.PersistentFlags().StringVar(&flagSignKeySignPubPath, "sign-pub-path", "", "path of signing public key")
	cmdSignKey.PersistentFlags().StringVar(&flagSignKeySigPath, "sig-path", "", "path of signature path")
}

func cmdSignKeyFunc(cmd *cobra.Command, args []string) error {
	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting sign-key command")

	rkRaw, err := os.ReadFile(flagSignKeyRootPrivPath)
	if err != nil {
		return err
	}
	rk, err := distsign.ParseRootKey(rkRaw)
	if err != nil {
		return err
	}

	bundle, err := os.ReadFile(flagSignKeySignPubPath)
	if err != nil {
		return err
	}
	sig, err := rk.SignSigningKeys(bundle)
	if err != nil {
		return err
	}

	if err := os.WriteFile(flagSignKeySigPath, sig, 0400); err != nil {
		return fmt.Errorf("failed writing signature file: %w", err)
	}
	fmt.Println("wrote signature to", flagSignKeySigPath)

	return nil
}
