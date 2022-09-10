package service

import (
	"context"
	"net"
)

// Check interface implementation
var (
	_ Security = (*SecurityService)(nil)
)

type SecurityService struct {
	TrustedSubnet *net.IPNet
}

func NewSecurityService(trustedSubnet *net.IPNet) *SecurityService {
	return &SecurityService{
		TrustedSubnet: trustedSubnet,
	}
}

func (s *SecurityService) IsIpAddrTrusted(ctx context.Context, ipStr string) (bool, error) {
	if s.TrustedSubnet == nil {
		return true, nil
	}

	ip, _, err := net.ParseCIDR(ipStr)
	if err != nil {
		return false, err
	}
	return s.TrustedSubnet.Contains(ip), nil
}
