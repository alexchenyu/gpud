package release

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/log"
	"github.com/leptonai/gpud/pkg/release/distsign"
)

var cmdVerifyKeySignature = &cobra.Command{
	Use:   "verify-key-signature",
	Short: "verify a root signture of the signing keys' bundle",
	RunE:  cmdVerifyKeySignatureFunc,
}

var (
	flagVerifyKeySignatureRootPubPath string
	flagVerifyKeySignatureSignPubPath string
	flagVerifyKeySignatureSigPath     string
)

func init() {
	cmdVerifyKeySignature.PersistentFlags().StringVar(&flagVerifyKeySignatureRootPubPath, "root-pub-path", "", "path of root public key")
	cmdVerifyKeySignature.PersistentFlags().StringVar(&flagVerifyKeySignatureSignPubPath, "sign-pub-path", "", "path of signing public key")
	cmdVerifyKeySignature.PersistentFlags().StringVar(&flagVerifyKeySignatureSigPath, "sig-path", "", "path of signature")
}

func cmdVerifyKeySignatureFunc(cmd *cobra.Command, args []string) error {
	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting verify-key-signature command")

	rootPubBundle, err := os.ReadFile(flagVerifyKeySignatureRootPubPath)
	if err != nil {
		return err
	}
	rootPubs, err := distsign.ParseRootKeyBundle(rootPubBundle)
	if err != nil {
		return fmt.Errorf("parsing %q: %w", flagVerifyKeySignatureRootPubPath, err)
	}

	signPubBundle, err := os.ReadFile(flagVerifyKeySignatureSignPubPath)
	if err != nil {
		return err
	}
	sig, err := os.ReadFile(flagVerifyKeySignatureSigPath)
	if err != nil {
		return err
	}

	if !distsign.VerifyAny(rootPubs, signPubBundle, sig) {
		return errors.New("signature not valid")
	}

	fmt.Println("signature ok")
	return nil
}
