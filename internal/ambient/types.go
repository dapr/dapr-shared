package ambient

// DaprTrustBundle represents all keys provided by Dapr.
type DaprTrustBundle struct {
	CertKey      string
	CertChain    string
	TrustAnchors string
}

// ToMap convert a DaprTrustBundle struct to map[string]string.
func (d *DaprTrustBundle) ToMap() map[string]string {
	return map[string]string{
		"dapr-trust-anchors": d.TrustAnchors,
		"dapr-cert-chain":    d.CertChain,
		"dapr-cert-key":      d.CertKey,
	}
}
