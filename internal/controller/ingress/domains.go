package ingress

import (
	"context"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/log"

	networkingv1alpha2 "github.com/0ekk/cloudflare-operator/api/v1alpha2"
)

type tunnelDomainMatch struct {
	Hostname string
	Domain   string
	ZoneID   string
}

func (r *Reconciler) matchTunnelDomain(
	ctx context.Context,
	config *networkingv1alpha2.TunnelIngressClassConfig,
	hostname string,
) (*tunnelDomainMatch, bool) {
	if hostname == "" {
		return &tunnelDomainMatch{Hostname: ""}, true
	}

	tunnel, err := r.getTunnel(ctx, config)
	if err != nil {
		log.FromContext(ctx).Info("Failed to resolve tunnel for hostname domain matching", "hostname", hostname, "error", err.Error())
		return nil, false
	}

	spec := tunnel.GetSpec()
	status := tunnel.GetStatus()
	hostname = strings.Trim(strings.ToLower(hostname), ".")

	configuredMatch, configuredHostname := matchConfiguredTunnelDomain(hostname, spec, status)
	if configuredMatch != nil {
		return configuredMatch, true
	}
	if configuredHostname {
		log.FromContext(ctx).Info("Skipping Ingress hostname", "hostname", hostname, "reason", "configured tunnel domain is not ready")
		return nil, false
	}

	legacy := defaultDomainMatch(spec, status)
	if legacy == nil {
		log.FromContext(ctx).Info("Skipping Ingress hostname", "hostname", hostname, "reason", "no default tunnel domain is available")
		return nil, false
	}
	legacy.Hostname = hostnameWithDefaultDomain(hostname, legacy.Domain)
	return legacy, true
}

func matchConfiguredTunnelDomain(
	hostname string,
	spec networkingv1alpha2.TunnelSpec,
	status networkingv1alpha2.TunnelStatus,
) (*tunnelDomainMatch, bool) {
	if len(spec.Cloudflare.Domains) == 0 {
		return nil, false
	}

	configuredDomains := allowedDomainSet(spec.Cloudflare.Domains)
	if best := bestDomainMatch(hostname, configuredDomainMatches(spec.Cloudflare.Domains, status.Domains)); best != nil {
		best.Hostname = hostname
		return best, true
	}
	return nil, hostnameMatchesConfiguredDomain(hostname, configuredDomains)
}

func bestDomainMatch(hostname string, allowed []tunnelDomainMatch) *tunnelDomainMatch {
	var best *tunnelDomainMatch
	for _, candidate := range allowed {
		if !candidate.matches(hostname) {
			continue
		}
		if best == nil || len(candidate.Domain) > len(best.Domain) {
			candidateCopy := candidate
			best = &candidateCopy
		}
	}
	return best
}

func (m tunnelDomainMatch) matches(hostname string) bool {
	if m.Domain == "" || m.ZoneID == "" {
		return false
	}
	return hostname == m.Domain || strings.HasSuffix(hostname, "."+m.Domain)
}

func configuredDomainMatches(configuredDomains []string, domainStatuses []networkingv1alpha2.TunnelDomainStatus) []tunnelDomainMatch {
	allowed := allowedDomainSet(configuredDomains)
	result := make([]tunnelDomainMatch, 0, len(domainStatuses))
	for _, domainStatus := range domainStatuses {
		normalized := normalizeDomain(domainStatus.Domain)
		if _, ok := allowed[normalized]; !ok || !isReadyDomainStatus(domainStatus) {
			continue
		}
		result = append(result, tunnelDomainMatch{Domain: normalized, ZoneID: domainStatus.ZoneId})
	}
	return result
}

func allowedDomainSet(domains []string) map[string]struct{} {
	allowed := make(map[string]struct{}, len(domains))
	for _, domain := range domains {
		if normalized := normalizeDomain(domain); normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}
	return allowed
}

func hostnameMatchesConfiguredDomain(hostname string, configuredDomains map[string]struct{}) bool {
	for domain := range configuredDomains {
		if (tunnelDomainMatch{Domain: domain, ZoneID: "configured"}).matches(hostname) {
			return true
		}
	}
	return false
}

func isReadyDomainStatus(domainStatus networkingv1alpha2.TunnelDomainStatus) bool {
	return domainStatus.State == networkingv1alpha2.TunnelDomainStateReady && domainStatus.ZoneId != ""
}

func defaultDomainMatch(
	spec networkingv1alpha2.TunnelSpec,
	status networkingv1alpha2.TunnelStatus,
) *tunnelDomainMatch {
	domain := normalizeDomain(spec.Cloudflare.Domain)
	zoneID := firstNonEmpty(spec.Cloudflare.ZoneId, status.ZoneId)
	if domain == "" || zoneID == "" {
		return nil
	}
	return &tunnelDomainMatch{
		Domain: domain,
		ZoneID: zoneID,
	}
}

func hostnameWithDefaultDomain(hostname, domain string) string {
	if (tunnelDomainMatch{Domain: domain, ZoneID: "default"}).matches(hostname) {
		return hostname
	}
	return hostname + "." + domain
}

func normalizeDomain(domain string) string {
	return strings.Trim(strings.ToLower(strings.TrimSpace(domain)), ".")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
