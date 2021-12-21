// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

import { User, Application, NamedApplicationFeatureFlag } from "./protocol";
import { ApplicationInstance, ApplicationInstancePhase } from "./application-instance";
import { RoleOrPermission } from "./permission";
import { AccountStatement } from "./accounting-protocol";

export interface AdminServer {
    adminGetUsers(req: AdminGetListRequest<User>): Promise<AdminGetListResult<User>>;
    adminGetUser(id: string): Promise<User>;
    adminBlockUser(req: AdminBlockUserRequest): Promise<User>;
    adminDeleteUser(id: string): Promise<void>;
    adminModifyRoleOrPermission(req: AdminModifyRoleOrPermissionRequest): Promise<User>;
    adminModifyPermanentApplicationFeatureFlag(req: AdminModifyPermanentApplicationFeatureFlagRequest): Promise<User>;

    adminGetApplications(req: AdminGetApplicationsRequest): Promise<AdminGetListResult<ApplicationAndInstance>>;
    adminGetApplication(id: string): Promise<ApplicationAndInstance>;
    adminForceStopApplication(id: string): Promise<void>;
    adminRestoreSoftDeletedApplication(id: string): Promise<void>;

    adminSetLicense(key: string): Promise<void>;

    adminGetAccountStatement(userId: string): Promise<AccountStatement>;
    adminSetProfessionalOpenSource(userId: string, shouldGetProfOSS: boolean): Promise<void>;
    adminIsStudent(userId: string): Promise<boolean>;
    adminAddStudentEmailDomain(userId: string, domain: string): Promise<void>;
    adminGrantExtraHours(userId: string, extraHours: number): Promise<void>;
}

export interface AdminGetListRequest<T> {
    offset: number
    limit: number
    orderBy: keyof T
    orderDir: "asc" | "desc"
    searchTerm?: string;
}

export interface AdminGetListResult<T> {
    total: number
    rows: T[]
}

export interface AdminBlockUserRequest {
    id: string
    blocked: boolean
}

export interface AdminModifyRoleOrPermissionRequest {
    id: string;
    rpp: {
        r: RoleOrPermission
        add: boolean
    }[]
}

export interface AdminModifyPermanentApplicationFeatureFlagRequest {
    id: string;
    changes: {
        featureFlag: NamedApplicationFeatureFlag
        add: boolean
    }[]
}

export interface ApplicationAndInstance extends Omit<Application, "id"|"creationTime">, Omit<ApplicationInstance, "id"|"creationTime"> {
    applicationId: string;
    applicationCreationTime: string;
    instanceId: string;
    instanceCreationTime: string;
    phase: ApplicationInstancePhase;
}

export namespace ApplicationAndInstance {
    export function toApplication(wai: ApplicationAndInstance): Application {
        return {
            id: wai.ApplicationId,
            creationTime: wai.applicationCreationTime,
            ... wai
        };
    }

    export function toInstance(wai: ApplicationAndInstance): ApplicationInstance | undefined {
        if (!wai.instanceId) {
            return undefined;
        }
        return {
            id: wai.instanceId,
            creationTime: wai.instanceCreationTime,
            ... wai
        };
    }
}

export type AdminGetApplicationsRequest = AdminGetListRequest<ApplicationAndInstance> & AdminGetApplicationsQuery;
/** The fields are meant to be used either OR (not combined) */
export type AdminGetApplicationsQuery = {
    /** we use this field in case we have a UUIDv4 and don't know whether it's an (old) application or instance id */
    instanceIdOrApplicationId?: string;
    instanceId?: string;
    applicationId?: string;
    ownerId?: string;
};
