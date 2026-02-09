import type { IExecuteFunctions, INodeExecutionData, INodeProperties } from 'n8n-workflow';
import * as asset from './asset';
import * as control from './control';
import * as datum from './datum';
import * as execute from './execute';
import * as framework from './framework';
import * as measure from './measure';
import * as member from './member';
import * as meeting from './meeting';
import * as organization from './organization';
import * as risk from './risk';
import * as vendor from './vendor';

export interface ResourceModule {
	description: INodeProperties[];
	[key: string]: OperationModule | INodeProperties[];
}

export interface OperationModule {
	description: INodeProperties[];
	execute: (this: IExecuteFunctions, itemIndex: number) => Promise<INodeExecutionData>;
}

export const resources: Record<string, ResourceModule> = {
	asset: asset as ResourceModule,
	control: control as ResourceModule,
	datum: datum as ResourceModule,
	execute: execute as ResourceModule,
	framework: framework as ResourceModule,
	measure: measure as ResourceModule,
	member: member as ResourceModule,
	meeting: meeting as ResourceModule,
	organization: organization as ResourceModule,
	risk: risk as ResourceModule,
	vendor: vendor as ResourceModule,
};

export function getAllResourceOperations(): INodeProperties[] {
	const operations: INodeProperties[] = [];

	for (const resource of Object.values(resources)) {
		const operationProp = resource.description.find((prop) => prop.name === 'operation');
		if (operationProp) {
			operations.push(operationProp);
		}
	}

	return operations;
}

export function getAllResourceFields(): INodeProperties[] {
	const fields: INodeProperties[] = [];

	for (const resource of Object.values(resources)) {
		fields.push(...resource.description.filter((prop) => prop.name !== 'operation'));
	}

	return fields;
}

export function getExecuteFunction(resourceName: string, operationName: string) {
	const resource = resources[resourceName];
	if (!resource) {
		throw new Error(`Unknown resource: ${resourceName}`);
	}

	const operationKey = resourceName === 'execute' ? 'execute' : operationName;

	const operation = resource[operationKey] as OperationModule;

	if (!operation || typeof operation.execute !== 'function') {
		throw new Error(`Unknown operation: ${operationName} for resource: ${resourceName}`);
	}

	return operation.execute;
}
