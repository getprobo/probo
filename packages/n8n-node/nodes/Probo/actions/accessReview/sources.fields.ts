// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

export const accessReviewSourceIdsField: INodeProperties = {
	displayName: 'Source Names or IDs',
	name: 'accessReviewSourceIds',
	type: 'multiOptions',
	displayOptions: {
		show: {
			resource: ['accessReview'],
			operation: ['create'],
		},
	},
	typeOptions: {
		loadOptionsMethod: 'getAccessReviewSources',
		loadOptionsDependsOn: ['organizationId'],
	},
	default: [],
	description: 'Scope sources to include in the campaign. Choose from the list, or specify IDs using an <a href="https://docs.n8n.io/code/expressions/">expression</a>.',
};

export const accessReviewSourceIdsUpdateField: INodeProperties = {
	displayName: 'Source Names or IDs',
	name: 'accessReviewSourceIds',
	type: 'multiOptions',
	typeOptions: {
		loadOptionsMethod: 'getAccessReviewSources',
		loadOptionsDependsOn: ['accessReviewCampaignId'],
	},
	default: [],
	description: 'Replace the campaign scope sources with this selection. Choose from the list, or specify IDs using an <a href="https://docs.n8n.io/code/expressions/">expression</a>.',
};
