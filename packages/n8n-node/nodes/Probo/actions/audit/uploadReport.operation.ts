import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiMultipartRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Report ID',
		name: 'reportId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['report'],
				operation: ['uploadReport'],
			},
		},
		default: '',
		description: 'The ID of the report to upload the file to',
		required: true,
	},
	{
		displayName: 'Input Data Field Name',
		name: 'binaryPropertyName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['report'],
				operation: ['uploadReport'],
			},
		},
		default: 'data',
		description: 'The name of the input field containing the binary file data to upload',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const reportId = this.getNodeParameter('reportId', itemIndex) as string;
	const binaryPropertyName = this.getNodeParameter('binaryPropertyName', itemIndex) as string;

	const binaryData = this.helpers.assertBinaryData(itemIndex, binaryPropertyName);
	const fileBuffer = await this.helpers.getBinaryDataBuffer(itemIndex, binaryPropertyName);

	const fileName = binaryData.fileName || 'report';
	const mimeType = binaryData.mimeType || 'application/octet-stream';

	const query = `
		mutation UploadReportFile($input: UploadReportFileInput!) {
			uploadReportFile(input: $input) {
				report {
					id
					name
					state
					validFrom
					validUntil
					reportUrl
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		input: {
			reportId,
			file: null,
		},
	};

	const responseData = await proboApiMultipartRequest.call(
		this,
		query,
		variables,
		'variables.input.file',
		fileBuffer,
		fileName,
		mimeType,
	);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
