package release

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/leptonai/gpud/cmd/gpud/common"
	"github.com/leptonai/gpud/pkg/log"
	"github.com/leptonai/gpud/pkg/release/distsign"
)

var cmdVerifyPackageSignature = &cobra.Command{
	Use:   "verify-package-signature",
	Short: "verify a package signture using a signing key",
	RunE:  cmdVerifyPackageSignatureFunc,
}

var (
	flagVerifyPackageSignaturePackagePath string
	flagVerifyPackageSignatureSignPubPath string
	flagVerifyPackageSignatureSigPath     string
)

func init() {
	cmdVerifyPackageSignature.PersistentFlags().StringVar(&flagVerifyPackageSignaturePackagePath, "package-path", "", "path of package")
	cmdVerifyPackageSignature.PersistentFlags().StringVar(&flagVerifyPackageSignatureSignPubPath, "sign-pub-path", "", "path of signing public key")
	cmdVerifyPackageSignature.PersistentFlags().StringVar(&flagVerifyPackageSignatureSigPath, "sig-path", "", "path of signature path")
}

func cmdVerifyPackageSignatureFunc(cmd *cobra.Command, args []string) error {
	var err error
	log.Logger, _, err = common.CreateLoggerFromFlags(cmd)
	if err != nil {
		return err
	}

	log.Logger.Debugw("starting verify-package-signature command")

	signPubBundle, err := os.ReadFile(flagVerifyPackageSignatureSignPubPath)
	if err != nil {
		return err
	}
	signPubs, err := distsign.ParseSigningKeyBundle(signPubBundle)
	if err != nil {
		return fmt.Errorf("parsing %q: %w", flagVerifyPackageSignatureSignPubPath, err)
	}

	pkg, err := os.Open(flagVerifyPackageSignaturePackagePath)
	if err != nil {
		return err
	}
	defer pkg.Close()

	pkgHash := distsign.NewPackageHash()
	if _, err := io.Copy(pkgHash, pkg); err != nil {
		return fmt.Errorf("reading %q: %w", flagVerifyPackageSignaturePackagePath, err)
	}

	hash := binary.LittleEndian.AppendUint64(pkgHash.Sum(nil), uint64(pkgHash.Len()))
	sig, err := os.ReadFile(flagVerifyPackageSignatureSigPath)
	if err != nil {
		return err
	}
	if !distsign.VerifyAny(signPubs, hash, sig) {
		return errors.New("signature not valid")
	}

	fmt.Println("signature ok")
	return nil
}
