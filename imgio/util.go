package imgio

import (
	"archive/tar"
	"os"

	om "github.com/box-builder/overmount"
	digest "github.com/opencontainers/go-digest"
)

const emptyDigest = digest.Digest("sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")

func calcLayer(parentDigest digest.Digest, iter *om.Layer, tf *os.File) (digest.Digest, digest.Digest, error) {
	packDigest, err := iter.Pack(tf)
	if err != nil {
		return "", "", err
	}

	var chainID digest.Digest
	if iter.Parent != nil {
		chainID = digest.FromBytes([]byte(parentDigest.Hex() + " " + string(packDigest.Hex())))
	} else {
		chainID = packDigest
	}

	return chainID, packDigest, nil
}

func runChain(layer *om.Layer, tw *tar.Writer, run func(digest.Digest, *om.Layer, *tar.Writer) (digest.Digest, digest.Digest, int64, error)) ([]digest.Digest, []digest.Digest, []int64, []*om.Layer, error) {
	layers := []*om.Layer{}
	chainIDs := []digest.Digest{}
	diffIDs := []digest.Digest{}
	sizes := []int64{}

	var parent digest.Digest

	// we need to walk it from the root up; so we need to reverse the list.
	for iter := layer; iter != nil; iter = iter.Parent {
		layers = append(layers, iter)
	}

	for i := len(layers) - 1; i >= 0; i-- {
		iter := layers[i]
		chainID, diffID, size, err := run(parent, iter, tw)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		chainIDs = append(chainIDs, chainID)
		diffIDs = append(diffIDs, diffID)
		sizes = append(sizes, size)

		parent = chainID
	}

	return chainIDs, diffIDs, sizes, layers, nil
}
