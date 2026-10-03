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

import type { INodeProperties } from 'n8n-workflow';

const cadences = [
	{ name: 'Not set', value: '' },
	{ name: 'Continuous', value: 'CONTINUOUS' },
	{ name: 'Daily', value: 'DAILY' },
	{ name: 'Weekly', value: 'WEEKLY' },
	{ name: 'Monthly', value: 'MONTHLY' },
	{ name: 'Quarterly', value: 'QUARTERLY' },
	{ name: 'Semiannually', value: 'SEMIANNUALLY' },
	{ name: 'Annually', value: 'ANNUALLY' },
	{ name: 'Ad hoc', value: 'AD_HOC' },
];

export function internalControlFields(operation: string): INodeProperties[] {
	const displayOptions = {
		show: {
			resource: ['measure'],
			operation: [operation],
		},
	};

	return [
		{
			displayName: 'Code',
			name: 'code',
			type: 'string',
			displayOptions,
			default: '',
			description: 'Stable reference, for example IC-ACCESS-01',
		},
		{
			displayName: 'Control Type',
			name: 'controlType',
			type: 'options',
			displayOptions,
			options: [
				{ name: 'Not set', value: '' },
				{ name: 'Preventive', value: 'PREVENTIVE' },
				{ name: 'Detective', value: 'DETECTIVE' },
				{ name: 'Corrective', value: 'CORRECTIVE' },
			],
			default: '',
			description: 'Whether the control prevents, detects, or corrects',
		},
		{
			displayName: 'Nature',
			name: 'nature',
			type: 'options',
			displayOptions,
			options: [
				{ name: 'Not set', value: '' },
				{ name: 'Manual', value: 'MANUAL' },
			],
			default: '',
			description: 'How the control is performed',
		},
		{
			displayName: 'Operating Frequency',
			name: 'operatingFrequency',
			type: 'options',
			displayOptions,
			options: cadences,
			default: '',
			description: 'How often the control runs',
		},
		{
			displayName: 'Evidence Cadence',
			name: 'evidenceCadence',
			type: 'options',
			displayOptions,
			options: cadences,
			default: '',
			description: 'How often evidence is collected',
		},
		{
			displayName: 'Testing Cadence',
			name: 'testingCadence',
			type: 'options',
			displayOptions,
			options: cadences,
			default: '',
			description: 'How often effectiveness is tested',
		},
		{
			displayName: 'Implementation Status',
			name: 'implementationStatus',
			type: 'options',
			displayOptions,
			options: [
				{ name: 'Not set', value: '' },
				{ name: 'Not implemented', value: 'NOT_IMPLEMENTED' },
				{ name: 'In progress', value: 'IN_PROGRESS' },
				{ name: 'Implemented', value: 'IMPLEMENTED' },
				{ name: 'Operating', value: 'OPERATING' },
			],
			default: '',
			description: 'Whether the control is in place and running',
		},
		{
			displayName: 'Owner ID',
			name: 'ownerId',
			type: 'string',
			displayOptions,
			default: '',
			description: 'Membership profile ID of the owner',
		},
		{
			displayName: 'Reviewer ID',
			name: 'reviewerId',
			type: 'string',
			displayOptions,
			default: '',
			description: 'Membership profile ID of the reviewer. Must differ from the owner.',
		},
	];
}

const fieldNames = [
	'code',
	'controlType',
	'nature',
	'operatingFrequency',
	'evidenceCadence',
	'testingCadence',
	'implementationStatus',
	'ownerId',
	'reviewerId',
] as const;

export function readInternalControlFields(
	get: (name: string) => string,
): Record<string, string> {
	const input: Record<string, string> = {};
	for (const name of fieldNames) {
		const value = get(name);
		if (value) {
			input[name] = value;
		}
	}

	return input;
}
