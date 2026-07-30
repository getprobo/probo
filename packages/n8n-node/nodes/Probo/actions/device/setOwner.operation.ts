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
		displayName: 'Device ID',
		name: 'deviceId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['device'],
				operation: ['setOwner'],
			},
		},
		default: '',
		description: 'The ID of the device',
		required: true,
	},
	{
		displayName: 'Clear Owner',
		name: 'clearOwner',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['device'],
				operation: ['setOwner'],
			},
		},
		default: false,
		description: 'Whether to clear the device owner instead of assigning one',
	},
	{
		displayName: 'Owner ID',
		name: 'ownerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['device'],
				operation: ['setOwner'],
				clearOwner: [false],
			},
		},
		default: '',
		description: 'The profile ID of the owner',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const deviceId = this.getNodeParameter('deviceId', itemIndex) as string;
	const clearOwner = this.getNodeParameter('clearOwner', itemIndex) as boolean;

	const input: { deviceId: string; ownerId: string | null } = {
		deviceId,
		ownerId: null,
	};

	if (!clearOwner) {
		input.ownerId = this.getNodeParameter('ownerId', itemIndex) as string;
	}

	const query = `
		mutation SetDeviceOwner($input: SetDeviceOwnerInput!) {
			setDeviceOwner(input: $input) {
				device {
					id
					owner {
						id
						fullName
					}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
