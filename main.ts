// Copyright (c) HashiCorp, Inc
// SPDX-License-Identifier: MPL-2.0
import { Construct } from "constructs";
import { App, TerraformStack, CloudBackend, NamedCloudWorkspace } from "cdktf";
import * as google from '@cdktf/provider-google';
import { AssetType, TerraformAsset } from "cdktf/lib/terraform-asset";
import * as path from 'path';

const project = 'qiita-gcp-cf-line-bot';
const region = 'us-central1';

class MyStack extends TerraformStack {
  constructor(scope: Construct, id: string) {
    super(scope, id);

    new google.provider.GoogleProvider(this, 'google', {
      project,
      region,
    });

    const functionRunner = new google.serviceAccount.ServiceAccount(this, 'functionRunner', {
      accountId: 'function-runner',
    });

    const allUsersRunInvoker = new google.dataGoogleIamPolicy.DataGoogleIamPolicy(this, 'allUsersRunInvoker', {
      binding: [{
        members: ['allUsers'],
        role: 'roles/run.invoker',
      }],
    });

    const assetBucket = new google.storageBucket.StorageBucket(this, 'assetBucket', {
      location: region,
      name: `asset-bucket-${project}`,
    });

    const simpleAsset = new TerraformAsset(this, 'simpleAsset', {
      path: path.resolve('simple'),
      type: AssetType.ARCHIVE,
    });

    const simpleObject = new google.storageBucketObject.StorageBucketObject(this, 'simpleObject', {
      bucket: assetBucket.name,
      name: simpleAsset.assetHash,
      source: simpleAsset.path,
    });

    const simpleFunction = new google.cloudfunctions2Function.Cloudfunctions2Function(this, 'simpleFunction', {
      name: 'simple',
      buildConfig: {
        entryPoint: 'simple',
        runtime: 'go119',
        source: {
          storageSource: {
            bucket: assetBucket.name,
            object: simpleObject.name,
          },
        },
      },
      location: region,
      serviceConfig: {
        minInstanceCount: 0,
        maxInstanceCount: 0,
        serviceAccountEmail: functionRunner.email,
      },
    });

    new google.cloudRunServiceIamPolicy.CloudRunServiceIamPolicy(this, 'simpleFunctionNoAuth', {
      location: region,
      policyData: allUsersRunInvoker.policyData,
      service: simpleFunction.name,
    });

  }
}

const app = new App();
const stack = new MyStack(app, "qiita-gcp-cf-line-bot");
new CloudBackend(stack, {
  hostname: "app.terraform.io",
  organization: "hsmtkkdefault",
  workspaces: new NamedCloudWorkspace("qiita-gcp-cf-line-bot")
});
app.synth();
