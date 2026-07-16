package svc

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	merrors "go-micro.dev/v4/errors"

	"github.com/opencloud-eu/opencloud/services/thumbnails/pkg/thumbnail"
)

// timocloud patch #4: honest 404 for thumbnail requests. The thumbnails grpc
// service returns http.StatusNotFound both for files that genuinely don't
// exist and for files whose mimetype it can't process, distinguishing the
// latter only via merrors.Error.Detail == thumbnail.UnsupportedFileTypeDetail.
// thumbnailNotFoundMsg is what the four webdav StatusNotFound branches
// (SpacesThumbnail, Thumbnail, PublicThumbnail, PublicThumbnailHead) use to
// turn that into an honest message instead of always claiming the file
// "could not be located".
var _ = Describe("thumbnailNotFoundMsg", func() {
	It("passes through the unsupported-file-type detail unchanged", func() {
		e := &merrors.Error{Detail: thumbnail.UnsupportedFileTypeDetail}
		Expect(thumbnailNotFoundMsg(e, "diagram.vsdx")).To(Equal(thumbnail.UnsupportedFileTypeDetail))
	})

	It("falls back to the generic not-located message for any other detail", func() {
		e := &merrors.Error{Detail: "could not stat file: not found"}
		Expect(thumbnailNotFoundMsg(e, "missing.png")).To(Equal(notFoundMsg("missing.png")))
	})

	It("falls back to the generic not-located message when detail is empty", func() {
		e := &merrors.Error{}
		Expect(thumbnailNotFoundMsg(e, "missing.png")).To(Equal("File with name missing.png could not be located"))
	})
})
