package build

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"k8s.io/release/pkg/release"
)

type ReleasePushBuild struct {
	Location string
}

var _ Stager = &ReleasePushBuild{}

// Stage stages the build to GCS using
// essentially release/push-build.sh --bucket=B --ci --gcs-suffix=S --noupdatelatest
func (rpb *ReleasePushBuild) Stage(version string) error {
	re := regexp.MustCompile(`^gs://([\w-]+)/(devel|ci)(/.*)?`)
	mat := re.FindStringSubmatch(rpb.Location)
	if mat == nil || len(mat) < 4 {
		return fmt.Errorf("invalid stage location: %v. Use gs://bucket/ci/optional-suffix", rpb.Location)
	}

	return errors.Wrap(
		release.NewPushBuild(&release.PushBuildOptions{
			Bucket:         mat[1],
			BuildDir:       release.BuildDir,
			GCSSuffix:      mat[3],
			AllowDup:       true,
			CI:             mat[2] == "ci",
			NoUpdateLatest: true,
		}).Push(),
		"stage via krel push",
	)
}
