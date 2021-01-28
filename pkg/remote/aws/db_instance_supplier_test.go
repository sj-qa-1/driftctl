package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/parallel"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
)

func TestDBInstanceSupplier_Resources(t *testing.T) {

	tests := []struct {
		test           string
		dirName        string
		instancesPages mocks.DescribeDBInstancesPagesOutput
		err            error
	}{
		{
			test:    "no dbs",
			dirName: "db_instance_empty",
			instancesPages: mocks.DescribeDBInstancesPagesOutput{
				{
					true,
					&rds.DescribeDBInstancesOutput{},
				},
			},
			err: nil,
		},
		{
			test:    "single db",
			dirName: "db_instance_single",
			instancesPages: mocks.DescribeDBInstancesPagesOutput{
				{
					true,
					&rds.DescribeDBInstancesOutput{
						DBInstances: []*rds.DBInstance{
							{
								DBInstanceIdentifier: awssdk.String("terraform-20201015115018309600000001"),
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			test:    "multiples mixed db",
			dirName: "db_instance_multiple",
			instancesPages: mocks.DescribeDBInstancesPagesOutput{
				{
					true,
					&rds.DescribeDBInstancesOutput{
						DBInstances: []*rds.DBInstance{
							{
								DBInstanceIdentifier: awssdk.String("terraform-20201015115018309600000001"),
							},
							{
								DBInstanceIdentifier: awssdk.String("database-1"),
							},
						},
					},
				},
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}
			providerLibrary.AddProvider(terraform.AWS, provider)
			supplierLibrary.AddSupplier(NewDBInstanceSupplier(provider))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewDBInstanceDeserializer()
			s := &DBInstanceSupplier{
				provider,
				deserializer,
				mocks.NewMockAWSRDSClient(tt.instancesPages),
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			if tt.err != err {
				t.Errorf("Expected error %+v got %+v", tt.err, err)
			}

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
