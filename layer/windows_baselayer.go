package layer

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
)

const encodedBytes = "H4sIAAAAAAAC/+ycy28bVRTGP6cpLSmUiR9RkLoYlAhY4HqetmcV8rCbQB4ocVJVymY8M66nediyp5SoCsoyyy7YsEDKgkU3dMuSbpAQC9Q/ITu6QCgSm8CCQXfGE8fxGIia2EacX2XP9X3Mufd85051r+Zm1bE3bWdnbQGXh+DT6dqeFkVBlMGr6AIP645egyBc9CD/I5zon8rbm1a9T/SXBFUl/XugfyqXn+sH/ZW0QPr3SP/Ugm3UKvVKyemZ/ooi0fzvvf6pqUrF6YX+qizT/O8T/VNT0zOv7pq0onTQXw7TPyMr4AXS/9KpWfdLkQjAPof7L7568+sfWRJXG+UszcP/Yul1TGAC61jFIqaxDgc12DCw4aW3UMU6ZpBDHpNYxTwKIPqV7+zsH+QFgiAIgiAIgiCI/wflor3tJbhmXrAPwNIHrutub3wA/r2pL4O859eAgUbdF1HAdV2Xpdl1r3EdCbH1BoDx8fFPlpcKS7dn59ZY3g9/um59A2Dt2OcGgFsRIILBLxABuIjfNg5m8xcMeb9G8TYGEPE7cZUH8A5LD42y33tt+SH1xzrUH2vUn8BbLfX5Zn2uLR9xv77f16tR9j0U9+uezut435a6V7x9lzN1Wu8DzN6k2CUIgiAIgiAIgiDOuf7nWtf/p3nqrf/5lvU/j+b6/zAWvv6fD7nXNQBLxQeW4dQ9c5vlCHgO+DX37KP9MDtc0w4/8u/3Gd4F8FgzLT2bloykaphmUrEyQlI3DDFZkqWirCimktHUXZz0Y58D7nzu3DkI6QcriwTjHQ7vR5j7rgPIbVpb1rZz8mJtYI8bBrRnPz8Ns8fKAtwGrOBo2N8jCewzbgJYPGVPkv2XaoLyI9d1mb29YeB4aO3Bseu6x8PAoeu6n25c89qORv3xRZr99XR/jDR0pGFARAkiikhChQYTCpIQIcKCjiQ0KJCRbaQESLC8lkWY0KAhi92T8Rx0iKez42W6v4y2j5f3dlF8bgCYsepGza46dsXfxnrpjXcA1Shgyt//FMQX88NxFHg/1ti38sbPYRDYY50YbIw/b9e2Huk1a6Fi2iXbMk/2xBr+YrZnY01/fWztLOpbllc+hWnMQDjzz+ef4vsofs74TutpQyyJxaSqmUpSFC09qSlyNqkpgmSl9aKpadndpj/KMWDlt4nf/y7Oy7HwuCvH23W43vBZJx2COD+O+TqwuKvGAz8Oen6/Ap4L/F7YqVqd4qMca86/cuLV5l/gj6BfT+Kd5+GTeLg/PkuEz8Nbp+yKkjcPlbPz8HoCGPvmtW+ZP/YSzbhibQ8S7fOQlRfxEDY2YaIK3XvXqo7bsFHp4K99Lrzf2ZHz6xj463nC91fw3GL9/3CkVc8BcO16vk7/xxIEQRAEQfQTl3fqs8n5z/8IopKh8z9d0z911942K4/qfaO/mBZE0r/r+qdWduqOtSVLvdZflmSJ9O+d/imjsl2y7/dMf0XJ0PO/9/qnZnL5ydX5wgW4pvP5z9Dzv2qGzn/2gf4rkwsXNP/Pqb8oC6R/H+ifm15dnivc67r+GTVD+veB/kv5wt3J5Vz39Vck0r8P9L+3Usgt9OD5r9Lff+gOfwUAAP//TAHkRgBOAAA="

func BaseLayerBytes() ([]byte, error) {
	gzipBytes, err := base64.StdEncoding.DecodeString(encodedBytes)
	if err != nil {
		return nil, err
	}

	gzipReader, err := gzip.NewReader(bytes.NewBuffer(gzipBytes))
	if err != nil {
		return nil, err
	}

	decodedBytes, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		return nil, err
	}

	return decodedBytes, nil
}
