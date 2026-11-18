// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package coredata

import (
	"strings"
)

func (s AiSystem) SearchText() string {
	return joinSearch(
		s.Name,
		derefSearch(s.Version),
		derefSearch(s.Source),
		derefSearch(s.Purpose),
		derefSearch(s.IntendedUseCases),
		derefSearch(s.AutonomyLevel),
		derefSearch(s.HumanOversightMechanism),
		derefSearch(s.KeyStakeholders),
		derefSearch(s.DataSourcesAndType),
		derefSearch(s.Notes),
	)
}

func (a Asset) SearchText() string {
	return joinSearch(a.Name, a.DataTypesStored)
}

func (a Audit) SearchText() string {
	return joinSearch(derefSearch(a.Name), derefSearch(a.Firm))
}

func (bf BusinessFunction) SearchText() string {
	return joinSearch(bf.Name, bf.ReferenceID, derefSearch(bf.ImpactTolerance), derefSearch(bf.Notes))
}

func (c Control) SearchText() string {
	return joinSearch(c.SectionTitle, c.Name, derefSearch(c.Description))
}

func (d Datum) SearchText() string {
	return d.Name
}

func (d Device) SearchTitle() string {
	for _, part := range []*string{d.Hostname, d.SerialNumber, d.HardwareUUID} {
		if part == nil || strings.TrimSpace(*part) == "" {
			continue
		}

		return *part
	}

	return ""
}

func (d Device) SearchText() string {
	platform := ""
	if d.Platform != nil {
		platform = d.Platform.String()
	}

	return joinSearch(
		derefSearch(d.Hostname),
		derefSearch(d.SerialNumber),
		derefSearch(d.HardwareUUID),
		derefSearch(d.OSVersion),
		derefSearch(d.AgentVersion),
		platform,
		string(d.Labels),
	)
}

func (p Document) SearchText() string {
	return p.Title
}

func (e Evidence) SearchText() string {
	return joinSearch(e.ReferenceID, derefSearch(e.Description), e.URL)
}

func (f Finding) SearchText() string {
	return joinSearch(
		f.ReferenceID,
		derefSearch(f.Description),
		derefSearch(f.Source),
		derefSearch(f.RootCause),
		derefSearch(f.CorrectiveAction),
		derefSearch(f.EffectivenessCheck),
	)
}

func (f Framework) SearchText() string {
	return joinSearch(f.Name, derefSearch(f.Description))
}

func (m InternalControl) SearchText() string {
	return joinSearch(m.Name, derefSearch(m.Code), derefSearch(m.OperatingEvent), derefSearch(m.Description))
}

func (o Obligation) SearchText() string {
	return joinSearch(
		derefSearch(o.Requirement),
		derefSearch(o.Area),
		derefSearch(o.Source),
		derefSearch(o.Regulator),
		derefSearch(o.ActionsToBeImplemented),
	)
}

func (p ProcessingActivity) SearchText() string {
	return joinSearch(
		p.Name,
		derefSearch(p.Purpose),
		derefSearch(p.DataSubjectCategory),
		derefSearch(p.PersonalDataCategory),
		derefSearch(p.Recipients),
		derefSearch(p.Location),
		derefSearch(p.RetentionPeriod),
		derefSearch(p.SecurityMeasures),
	)
}

func (rr RightsRequest) SearchText() string {
	return joinSearch(
		derefSearch(rr.DataSubject),
		derefSearch(rr.Contact),
		derefSearch(rr.Details),
		derefSearch(rr.ActionTaken),
	)
}

func (r Risk) SearchText() string {
	return joinSearch(r.ReferenceID, r.Name)
}

func (ra RiskAnalysis) SearchText() string {
	return joinSearch(ra.Name, derefSearch(ra.Description))
}

func (s StatementOfApplicability) SearchText() string {
	return s.Name
}

func (t Task) SearchText() string {
	return joinSearch(t.Name, t.ReferenceID, t.Content)
}

func (v ThirdParty) SearchText() string {
	return joinSearch(v.Name, derefSearch(v.Description), derefSearch(v.LegalName))
}

func (d DataProtectionImpactAssessment) SearchText(activityName string) string {
	return joinSearch(
		activityName,
		derefSearch(d.Description),
		derefSearch(d.NecessityAndProportionality),
		derefSearch(d.PotentialRisk),
		derefSearch(d.Mitigations),
	)
}

func (t TransferImpactAssessment) SearchText(activityName string) string {
	return joinSearch(
		activityName,
		derefSearch(t.DataSubjects),
		derefSearch(t.LegalMechanism),
		derefSearch(t.Transfer),
		derefSearch(t.LocalLawRisk),
		derefSearch(t.SupplementaryMeasures),
	)
}

func (tp TreatmentPlan) SearchText(riskName string) string {
	return joinSearch(riskName, tp.Treatment.String())
}
