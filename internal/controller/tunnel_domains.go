package controller

import (
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkingv1alpha2 "github.com/0ekk/cloudflare-operator/api/v1alpha2"
	"github.com/0ekk/cloudflare-operator/internal/clients/cf"
)

const conditionDomainsReady = "DomainsReady"

type tunnelDomainResolution struct {
	Domain string
	ZoneID string
	Err    error
}

func resolveTunnelDomains(r GenericTunnelReconciler) []tunnelDomainResolution {
	domains := normalizeTunnelDomains(r.GetTunnel().GetSpec().Cloudflare.Domains)
	results := make([]tunnelDomainResolution, 0, len(domains))
	for _, domain := range domains {
		zoneID, _, err := r.GetCfAPI().GetZoneIDForDomain(r.GetContext(), domain)
		results = append(results, tunnelDomainResolution{
			Domain: domain,
			ZoneID: zoneID,
			Err:    err,
		})
	}
	return results
}

func normalizeTunnelDomains(domains []string) []string {
	seen := make(map[string]struct{}, len(domains))
	result := make([]string, 0, len(domains))
	for _, domain := range domains {
		normalized := strings.ToLower(strings.Trim(strings.TrimSpace(domain), "."))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

func applyTunnelDomainStatuses(status *networkingv1alpha2.TunnelStatus, resolutions []tunnelDomainResolution, generation int64) {
	if len(resolutions) == 0 {
		status.Domains = nil
		meta.SetStatusCondition(&status.Conditions, metav1.Condition{
			Type:               conditionDomainsReady,
			Status:             metav1.ConditionTrue,
			Reason:             "NoDomainsConfigured",
			Message:            "No additional tunnel domains configured",
			ObservedGeneration: generation,
		})
		return
	}

	now := metav1.Now()
	domainStatuses := make([]networkingv1alpha2.TunnelDomainStatus, 0, len(resolutions))
	var failures []string
	for _, resolution := range resolutions {
		domainStatus := networkingv1alpha2.TunnelDomainStatus{
			Domain:           resolution.Domain,
			ZoneId:           resolution.ZoneID,
			LastResolvedTime: &now,
		}
		if resolution.Err != nil {
			domainStatus.State = networkingv1alpha2.TunnelDomainStateError
			domainStatus.Message = cf.SanitizeErrorMessage(resolution.Err)
			failures = append(failures, fmt.Sprintf("%s: %s", resolution.Domain, domainStatus.Message))
		} else {
			domainStatus.State = networkingv1alpha2.TunnelDomainStateReady
			domainStatus.Message = "Domain resolved"
		}
		domainStatuses = append(domainStatuses, domainStatus)
	}
	status.Domains = domainStatuses

	if len(failures) > 0 {
		meta.SetStatusCondition(&status.Conditions, metav1.Condition{
			Type:               conditionDomainsReady,
			Status:             metav1.ConditionFalse,
			Reason:             "DomainResolutionFailed",
			Message:            strings.Join(failures, "; "),
			ObservedGeneration: generation,
		})
		return
	}

	meta.SetStatusCondition(&status.Conditions, metav1.Condition{
		Type:               conditionDomainsReady,
		Status:             metav1.ConditionTrue,
		Reason:             "Resolved",
		Message:            "All tunnel domains resolved",
		ObservedGeneration: generation,
	})
}
