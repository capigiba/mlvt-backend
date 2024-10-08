package entity

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type MoMoRequest struct {
	PartnerCode string
	AccessKey   string
	RequestID   string
	Amount      string
	OrderID     string
	Signature   string
}

// NewMoMoRequest is a factory function to create a MoMoRequest with the necessary fields
func NewMoMoRequest(partnerCode, accessKey, requestID, amount, orderID string) *MoMoRequest {
	return &MoMoRequest{
		PartnerCode: partnerCode,
		AccessKey:   accessKey,
		RequestID:   requestID,
		Amount:      amount,
		OrderID:     orderID,
	}
}

// GenerateSignature generates an HMAC SHA256 signature for the MoMo payment
func (r *MoMoRequest) GenerateSignature(secrectKey string) {
	rawSignature := r.PartnerCode + r.RequestID + r.Amount + r.AccessKey
	h := hmac.New(sha256.New, []byte(secrectKey))
	h.Write([]byte(rawSignature))
	r.Signature = hex.EncodeToString(h.Sum(nil))
}
