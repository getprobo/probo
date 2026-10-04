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
			displayName: 'Operating Mode',
			name: 'operatingMode',
			type: 'options',
			displayOptions,
			options: [
				{ name: 'Not set', value: '' },
				{ name: 'Continuous', value: 'CONTINUOUS' },
				{ name: 'When an event occurs', value: 'EVENT' },
				{ name: 'Every', value: 'PERIODIC' },
			],
			default: '',
			description: 'Whether the control runs always, when an event occurs, or on a duration',
		},
		{
			displayName: 'Operating Interval',
			name: 'operatingInterval',
			type: 'string',
			displayOptions: {
				show: {
					resource: ['measure'],
					operation: [operation],
					operatingMode: ['PERIODIC'],
				},
			},
			default: '',
			description: 'ISO-8601 duration for a periodic control, for example P3M',
		},
		{
			displayName: 'Operating Event',
			name: 'operatingEvent',
			type: 'string',
			displayOptions: {
				show: {
					resource: ['measure'],
					operation: [operation],
					operatingMode: ['EVENT'],
				},
			},
			default: '',
			description: 'What triggers an event-driven control, for example when someone leaves',
		},
		{
			displayName: 'Evidence Cadence',
			name: 'evidenceCadence',
			type: 'string',
			displayOptions,
			default: '',
			description: 'ISO-8601 duration, for example P1M',
		},
		{
			displayName: 'Testing Cadence',
			name: 'testingCadence',
			type: 'string',
			displayOptions,
			default: '',
			description: 'ISO-8601 duration, for example P3M',
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
	'evidenceCadence',
	'testingCadence',
	'implementationStatus',
	'ownerId',
	'reviewerId',
] as const;

export function readInternalControlFields(
	get: (name: string) => string,
): Record<string, unknown> {
	const input: Record<string, unknown> = {};
	for (const name of fieldNames) {
		const value = get(name);
		if (value) {
			input[name] = value;
		}
	}

	const mode = get('operatingMode');
	if (mode) {
		const frequency: Record<string, string> = { mode };
		if (mode === 'PERIODIC') {
			const interval = get('operatingInterval');
			if (interval) {
				frequency.interval = interval;
			}
		}
		if (mode === 'EVENT') {
			const event = get('operatingEvent');
			if (event) {
				frequency.event = event;
			}
		}
		input.operatingFrequency = frequency;
	}

	return input;
}
