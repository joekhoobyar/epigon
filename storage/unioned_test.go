package storage_test

import (
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github/joekhoobyar/epigon/storage"
	"github/joekhoobyar/epigon/test"
)

var _ = Describe("UnionedCache", func() {
	var root, child1, child2 []byte
	var err error

	dir := test.FixtureDir()
	u := storage.NewUnionedCache(dir)
	var c storage.RWCache = u // force breakage if we fail to implement the interface

	BeforeEach(func() {
		c.Clear()

		root, err = os.ReadFile(path.Join(dir, "root.json"))
		Expect(err).NotTo(HaveOccurred())
		child1, err = os.ReadFile(path.Join(dir, "root/child1.json"))
		Expect(err).NotTo(HaveOccurred())
		child2, err = os.ReadFile(path.Join(dir, "root/child2.json"))
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Reading locations", func() {

		Context("that are objects", func() {
			It("should load root", func() { Expect(u.Read("root")).To(Equal(root)) })
			It("should load child1", func() { Expect(u.Read("root/child1")).To(Equal(child1)) })
			It("should load child2", func() { Expect(u.Read("root/child2")).To(Equal(child2)) })
		})

		Context("that are lists", func() {
			It("should report an error", func() {
				_, err := u.Read("root/")
				Expect(err).To(MatchError(HaveSuffix(" location does not identify an object")))
			})
		})

		Context("that do not exist", func() {
			It("should report an error", func() {
				_, err = u.Read("missing")
				Expect(err).To(MatchError(HaveSuffix(" no such file or directory")))
			})
		})

		Context("that are hidden by holes", func() {
			It("should report an error", func() {
				Expect(u.Delete("root")).To(BeTrue())
				_, err = u.Read("root")
				Expect(err).To(MatchError(HaveSuffix(" no such object record")))
			})
		})
	})

	Describe("Reading fixture lists", func() {
		Context("that are objects", func() {
			It("should report an error", func() {
				_, err = u.ReadList("root/child1")
				Expect(err).To(MatchError(HaveSuffix(" location does not identify a collection")))
				_, err = u.ReadList("root")
				Expect(err).To(MatchError(HaveSuffix(" location does not identify a collection")))
			})
		})

		Context("that are lists", func() {
			It("should load as an array of immediate children", func() {
				Expect(u.ReadList("root/")).To(Equal([]byte("[{\"name\":\"baby\"},{\"name\":\"kid\"}]")))
			})
			It("should combine children from both layers", func() {
				err = u.Write("root/stepchild", []byte("{\"name\":\"headed\"}"))
				Expect(err).NotTo(HaveOccurred())
				Expect(u.ReadList("root/")).To(Equal([]byte("[{\"name\":\"baby\"},{\"name\":\"kid\"},{\"name\":\"headed\"}]")))
			})
			It("should load empty lists", func() {
				Expect(u.ReadList("root/child2/nest/")).To(Equal([]byte("[]")))
			})
		})
	})

	Describe("Listing fixture keys", func() {
		Context("that are objects", func() {
			It("should report an error", func() {
				_, err = u.List("root/child1")
				Expect(err).To(MatchError(HaveSuffix(" location does not identify a collection")))
				_, err = u.List("root")
				Expect(err).To(MatchError(HaveSuffix(" location does not identify a collection")))
			})
		})

		Context("that are lists", func() {
			It("should list subkeys of immediate children", func() {
				Expect(u.List("root/")).To(Equal([]string{"root/child1", "root/child2"}))
			})
			It("should combine children from both layers", func() {
				err = u.Write("root/stepchild", []byte("{\"name\":\"headed\"}"))
				Expect(err).NotTo(HaveOccurred())
				Expect(u.List("root/")).To(Equal([]string{"root/child1", "root/child2", "root/stepchild"}))
			})
			It("should load empty lists", func() {
				Expect(u.List("root/child2/nest/")).To(Equal([]string{}))
			})
		})
	})

	Describe("Writing locations", func() {
		Context("that are objects", func() {
			It("should be readable", func() {
				expected := []byte("\"guy\"")
				err := u.Write("other", expected)
				Expect(err).NotTo(HaveOccurred())
				actual, err := u.Read("other")
				Expect(actual).To(Equal(expected))
			})
			It("should hide read-only content", func() {
				expected := []byte("\"guy\"")
				err := u.Write("child1", expected)
				Expect(err).NotTo(HaveOccurred())
				actual, err := u.Read("child1")
				Expect(actual).To(Equal(expected))
			})
			It("should be listable", func() {
				err := u.Write("root/other", []byte("\"guy\""))
				Expect(err).NotTo(HaveOccurred())
				Expect(u.List("root/")).To(Equal([]string{"root/child1", "root/child2", "root/other"}))
			})
			It("should be readable as a JSON array", func() {
				err := u.Write("root/other", []byte("\"guy\""))
				Expect(err).NotTo(HaveOccurred())
				Expect(u.ReadList("root/")).To(Equal([]byte("[{\"name\":\"baby\"},{\"name\":\"kid\"},\"guy\"]")))
			})
		})

		Context("that are lists", func() {
			It("should report an error", func() {
				err = u.Write("root/", []byte("\"a\""))
				Expect(err).To(MatchError(HaveSuffix(" location does not identify an object")))
				err = u.Write("other/", []byte("\"a\""))
				Expect(err).To(MatchError(HaveSuffix(" location does not identify an object")))
			})
		})
	})

	Describe("Deleting locations", func() {
		Context("that are objects", func() {
			It("should succeed", func() {
				expected := []byte("\"guy\"")
				err := u.Write("other", expected)
				Expect(err).NotTo(HaveOccurred())
				Expect(u.Delete("other")).To(BeTrue())
				Expect(u.Exists("other")).To(BeFalse())
			})
		})

		Context("that are read-only", func() {
			It("should hide the object with a hole, and succeed", func() {
				Expect(u.Delete("root/child2")).To(BeTrue())
				Expect(u.Exists("root/child2")).To(BeFalse())
			})
		})

		Context("that are missing", func() {
			It("should fail", func() {
				Expect(u.Delete("missing")).To(BeFalse())
			})
		})

		Context("that are lists", func() {
			It("should fail", func() {
				Expect(u.Delete("root/")).To(BeFalse())
			})
		})
	})
})
