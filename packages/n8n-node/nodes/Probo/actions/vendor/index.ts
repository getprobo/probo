import type { INodeProperties } from 'n8n-workflow';
import * as createOp from './create.operation';
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as createServiceOp from './createService.operation';
import * as updateServiceOp from './updateService.operation';
import * as deleteServiceOp from './deleteService.operation';
import * as getServiceOp from './getService.operation';
import * as getAllServicesOp from './getAllServices.operation';
import * as createRiskAssessmentOp from './createRiskAssessment.operation';
import * as getRiskAssessmentOp from './getRiskAssessment.operation';
import * as getAllRiskAssessmentsOp from './getAllRiskAssessments.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['vendor'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new vendor',
				action: 'Create a vendor',
			},
			{
				name: 'Create Risk Assessment',
				value: 'createRiskAssessment',
				description: 'Create a new vendor risk assessment',
				action: 'Create a vendor risk assessment',
			},
			{
				name: 'Create Service',
				value: 'createService',
				description: 'Create a new vendor service',
				action: 'Create a vendor service',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a vendor',
				action: 'Delete a vendor',
			},
			{
				name: 'Delete Service',
				value: 'deleteService',
				description: 'Delete a vendor service',
				action: 'Delete a vendor service',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a vendor',
				action: 'Get a vendor',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many vendors',
				action: 'Get many vendors',
			},
			{
				name: 'Get Many Risk Assessments',
				value: 'getAllRiskAssessments',
				description: 'Get many vendor risk assessments',
				action: 'Get many vendor risk assessments',
			},
			{
				name: 'Get Many Services',
				value: 'getAllServices',
				description: 'Get many vendor services',
				action: 'Get many vendor services',
			},
			{
				name: 'Get Risk Assessment',
				value: 'getRiskAssessment',
				description: 'Get a vendor risk assessment',
				action: 'Get a vendor risk assessment',
			},
			{
				name: 'Get Service',
				value: 'getService',
				description: 'Get a vendor service',
				action: 'Get a vendor service',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing vendor',
				action: 'Update a vendor',
			},
			{
				name: 'Update Service',
				value: 'updateService',
				description: 'Update an existing vendor service',
				action: 'Update a vendor service',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...updateOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...createServiceOp.description,
	...updateServiceOp.description,
	...deleteServiceOp.description,
	...getServiceOp.description,
	...getAllServicesOp.description,
	...createRiskAssessmentOp.description,
	...getRiskAssessmentOp.description,
	...getAllRiskAssessmentsOp.description,
];

export {
	createOp as create,
	updateOp as update,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	createServiceOp as createService,
	updateServiceOp as updateService,
	deleteServiceOp as deleteService,
	getServiceOp as getService,
	getAllServicesOp as getAllServices,
	createRiskAssessmentOp as createRiskAssessment,
	getRiskAssessmentOp as getRiskAssessment,
	getAllRiskAssessmentsOp as getAllRiskAssessments,
};
