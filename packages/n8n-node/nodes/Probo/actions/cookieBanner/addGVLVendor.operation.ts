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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Cookie Banner ID',
		name: 'cookieBannerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['addGVLVendor'],
			},
		},
		default: '',
		description: 'The ID of the cookie banner',
		required: true,
	},
	{
		displayName: 'IAB Vendor ID',
		name: 'iabVendorId',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['addGVLVendor'],
			},
		},
		default: 0,
		description: 'The IAB GVL vendor ID to add',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const cookieBannerId = this.getNodeParameter('cookieBannerId', itemIndex) as string;
	const iabVendorId = this.getNodeParameter('iabVendorId', itemIndex) as number;

	const query = `
		mutation AddCookieBannerGVLVendor($input: AddCookieBannerGVLVendorInput!) {
			addCookieBannerGVLVendor(input: $input) {
				commonGVLVendor {
					id
					iabVendorId
					name
					policyUrl
				}
				cookieBanner {
					id
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, {
		input: { cookieBannerId, iabVendorId },
	});

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
