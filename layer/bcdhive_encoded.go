//+build !hivex

package layer

func BaseLayerBCD() ([]byte, error) {
	return DecodeBaseLayerBCD()
}
