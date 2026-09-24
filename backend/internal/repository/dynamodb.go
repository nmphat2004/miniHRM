package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"mini-hrm-backend/internal/model"
)

var (
	ErrNotFound        = errors.New("không tìm thấy bản ghi (404)")
	ErrDuplicateCode   = errors.New("mã đã tồn tại trong hệ thống")
	ErrDuplicateEmail  = errors.New("email đã được đăng ký trong hệ thống")
	ErrVersionConflict = errors.New("xung đột phiên bản (409 Conflict) - vui lòng tải lại dữ liệu")
	ErrIdempotencyHit  = errors.New("yêu cầu đã được xử lý (Idempotency)")
)

type DynamoRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoRepository(ctx context.Context, endpoint, region, tableName string) (*DynamoRepository, error) {
	var cfg aws.Config
	var err error

	if endpoint != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, reg string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           endpoint,
				SigningRegion: region,
			}, nil
		})
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithEndpointResolverWithOptions(customResolver),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "")),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
	}
	if err != nil {
		return nil, err
	}

	repo := &DynamoRepository{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}

	_ = repo.EnsureTable(ctx)
	return repo, nil
}

func (r *DynamoRepository) EnsureTable(ctx context.Context) error {
	_, err := r.client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(r.tableName),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("PK"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("SK"), KeyType: types.KeyTypeRange},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("SK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("GSI1PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("GSI1SK"), AttributeType: types.ScalarAttributeTypeS},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("GSI1"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("GSI1PK"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("GSI1SK"), KeyType: types.KeyTypeRange},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil && !strings.Contains(err.Error(), "ResourceInUseException") {
		return err
	}
	_, _ = r.client.UpdateTimeToLive(ctx, &dynamodb.UpdateTimeToLiveInput{
		TableName:               aws.String(r.tableName),
		TimeToLiveSpecification: &types.TimeToLiveSpecification{AttributeName: aws.String("expiresAt"), Enabled: aws.Bool(true)},
	})
	return nil
}

func (r *DynamoRepository) CheckIdempotency(ctx context.Context, key string) (bool, int, []byte, error) {
	if key == "" {
		return false, 0, nil, nil
	}
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("IDEMPOTENCY#%s", key)},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	})
	if err != nil {
		return false, 0, nil, err
	}
	if len(out.Item) == 0 {
		return false, 0, nil, nil
	}
	if expiry, ok := out.Item["expiresAt"].(*types.AttributeValueMemberN); ok {
		var expires int64
		_, _ = fmt.Sscan(expiry.Value, &expires)
		if expires > 0 && expires <= time.Now().Unix() {
			return false, 0, nil, nil
		}
	}

	statusCode := 200
	if sc, ok := out.Item["statusCode"].(*types.AttributeValueMemberN); ok {
		var code int
		fmt.Sscanf(sc.Value, "%d", &code)
		if code > 0 {
			statusCode = code
		}
	}
	var respBody []byte
	if rb, ok := out.Item["response"].(*types.AttributeValueMemberS); ok {
		respBody = []byte(rb.Value)
	}

	return true, statusCode, respBody, nil
}

func (r *DynamoRepository) PutDepartmentAtomic(ctx context.Context, dept *model.Department, audit *model.AuditLog, idempotencyKey string) error {
	deptPK := fmt.Sprintf("DEPT#%s", dept.ID)
	codePK := fmt.Sprintf("DEPT_CODE#%s", dept.Code)
	auditPK := fmt.Sprintf("AUDIT#%s", dept.ID)
	auditSK := fmt.Sprintf("LOG#%s#%s", audit.OccurredAt.Format(time.RFC3339Nano), audit.ID)

	parentGSI := "ROOT"
	if dept.ParentID != "" {
		parentGSI = dept.ParentID
	}

	deptItem, err := attributevalue.MarshalMap(dept)
	if err != nil {
		return err
	}
	deptItem["PK"] = &types.AttributeValueMemberS{Value: deptPK}
	deptItem["SK"] = &types.AttributeValueMemberS{Value: "METADATA"}
	deptItem["GSI1PK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPT_PARENT#%s", parentGSI)}
	deptItem["GSI1SK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("PATH#%s", dept.Path)}

	auditItem, err := attributevalue.MarshalMap(audit)
	if err != nil {
		return err
	}
	auditItem["PK"] = &types.AttributeValueMemberS{Value: auditPK}
	auditItem["SK"] = &types.AttributeValueMemberS{Value: auditSK}
	auditItem["GSI1PK"] = &types.AttributeValueMemberS{Value: "AUDIT_ALL"}
	auditItem["GSI1SK"] = &types.AttributeValueMemberS{Value: auditSK}

	actions := []types.TransactWriteItem{
		{
			Put: &types.Put{
				TableName:           aws.String(r.tableName),
				Item:                deptItem,
				ConditionExpression: aws.String("attribute_not_exists(PK)"),
			},
		},
		{
			Put: &types.Put{
				TableName: aws.String(r.tableName),
				Item: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: codePK},
					"SK": &types.AttributeValueMemberS{Value: "METADATA"},
				},
				ConditionExpression: aws.String("attribute_not_exists(PK)"),
			},
		},
		{
			Put: &types.Put{
				TableName: aws.String(r.tableName),
				Item:      auditItem,
			},
		},
	}
	if idempotencyKey != "" {
		response, _ := json.Marshal(map[string]interface{}{"success": true, "data": dept})
		actions = append(actions, idempotencyPut(r.tableName, idempotencyKey, http.StatusCreated, response))
	}

	_, err = r.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: actions,
	})
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailed") || strings.Contains(err.Error(), "TransactionCanceledException") {
			return ErrDuplicateCode
		}
		return err
	}
	return nil
}

func (r *DynamoRepository) GetDepartmentByID(ctx context.Context, id string) (*model.Department, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPT#%s", id)},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, ErrNotFound
	}

	var dept model.Department
	err = attributevalue.UnmarshalMap(out.Item, &dept)
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *DynamoRepository) GetAllDepartments(ctx context.Context) ([]*model.Department, error) {
	depts, err := r.ListDepartmentTree(ctx, "")
	if err != nil {
		return nil, err
	}

	sort.Slice(depts, func(i, j int) bool {
		return depts[i].Path < depts[j].Path
	})

	return depts, nil
}

// ListDepartmentTree walks from a root or a scoped department using the parent GSI.
// Each read is constrained in DynamoDB; it never scans unrelated departments.
func (r *DynamoRepository) ListDepartmentTree(ctx context.Context, rootID string) ([]*model.Department, error) {
	var result []*model.Department
	queue := []string{rootID}
	seen := map[string]bool{}
	for len(queue) > 0 {
		parentID := queue[0]
		queue = queue[1:]
		items, err := r.QueryDepartmentsByParent(ctx, parentID)
		if err != nil {
			return nil, err
		}
		for _, dept := range items {
			if seen[dept.ID] {
				continue
			}
			seen[dept.ID] = true
			result = append(result, dept)
			queue = append(queue, dept.ID)
		}
	}
	return result, nil
}

func (r *DynamoRepository) QueryDepartmentsByParent(ctx context.Context, parentID string) ([]*model.Department, error) {
	parent := parentID
	if parent == "" {
		parent = "ROOT"
	}
	var items []map[string]types.AttributeValue
	var start map[string]types.AttributeValue
	for {
		out, err := r.client.Query(ctx, &dynamodb.QueryInput{
			TableName: aws.String(r.tableName), IndexName: aws.String("GSI1"),
			KeyConditionExpression:    aws.String("GSI1PK = :parent"),
			ExpressionAttributeValues: map[string]types.AttributeValue{":parent": &types.AttributeValueMemberS{Value: "DEPT_PARENT#" + parent}},
			ExclusiveStartKey:         start,
		})
		if err != nil {
			return nil, err
		}
		items = append(items, out.Items...)
		if len(out.LastEvaluatedKey) == 0 {
			break
		}
		start = out.LastEvaluatedKey
	}
	var departments []*model.Department
	for _, item := range items {
		var dept model.Department
		if err := attributevalue.UnmarshalMap(item, &dept); err != nil {
			return nil, err
		}
		departments = append(departments, &dept)
	}
	return departments, nil
}

func (r *DynamoRepository) QueryEmployeesByDepartment(ctx context.Context, departmentID string) ([]*model.Employee, error) {
	var items []map[string]types.AttributeValue
	var start map[string]types.AttributeValue
	for {
		out, err := r.client.Query(ctx, &dynamodb.QueryInput{
			TableName: aws.String(r.tableName), IndexName: aws.String("GSI1"),
			KeyConditionExpression:    aws.String("GSI1PK = :dept"),
			ExpressionAttributeValues: map[string]types.AttributeValue{":dept": &types.AttributeValueMemberS{Value: "DEPT#" + departmentID}},
			ExclusiveStartKey:         start,
		})
		if err != nil {
			return nil, err
		}
		items = append(items, out.Items...)
		if len(out.LastEvaluatedKey) == 0 {
			break
		}
		start = out.LastEvaluatedKey
	}
	var employees []*model.Employee
	for _, item := range items {
		var employee model.Employee
		if err := attributevalue.UnmarshalMap(item, &employee); err != nil {
			return nil, err
		}
		employees = append(employees, &employee)
	}
	return employees, nil
}

func (r *DynamoRepository) UpdateDepartmentAtomic(ctx context.Context, dept *model.Department, audit *model.AuditLog, idempotencyKey string) error {
	prevVersion := dept.Version
	dept.Version++
	dept.UpdatedAt = time.Now().UTC()

	auditPK := fmt.Sprintf("AUDIT#%s", dept.ID)
	auditSK := fmt.Sprintf("LOG#%s#%s", audit.OccurredAt.Format(time.RFC3339Nano), audit.ID)

	parentGSI := "ROOT"
	if dept.ParentID != "" {
		parentGSI = dept.ParentID
	}

	deptItem, err := attributevalue.MarshalMap(dept)
	if err != nil {
		return err
	}
	deptItem["PK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPT#%s", dept.ID)}
	deptItem["SK"] = &types.AttributeValueMemberS{Value: "METADATA"}
	deptItem["GSI1PK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPT_PARENT#%s", parentGSI)}
	deptItem["GSI1SK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("PATH#%s", dept.Path)}

	auditItem, _ := attributevalue.MarshalMap(audit)
	auditItem["PK"] = &types.AttributeValueMemberS{Value: auditPK}
	auditItem["SK"] = &types.AttributeValueMemberS{Value: auditSK}
	auditItem["GSI1PK"] = &types.AttributeValueMemberS{Value: "AUDIT_ALL"}
	auditItem["GSI1SK"] = &types.AttributeValueMemberS{Value: auditSK}

	actions := []types.TransactWriteItem{
		{
			Put: &types.Put{
				TableName:           aws.String(r.tableName),
				Item:                deptItem,
				ConditionExpression: aws.String("version = :oldVer"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":oldVer": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", prevVersion)},
				},
			},
		},
		{
			Put: &types.Put{
				TableName: aws.String(r.tableName),
				Item:      auditItem,
			},
		},
	}
	if idempotencyKey != "" {
		actions = append(actions, idempotencyMessagePut(r.tableName, idempotencyKey, "Thao tác thành công"))
	}

	_, err = r.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: actions})
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailed") || strings.Contains(err.Error(), "TransactionCanceledException") {
			return ErrVersionConflict
		}
		return err
	}
	return nil
}

func (r *DynamoRepository) MoveDepartmentAtomic(ctx context.Context, dept *model.Department, subDepts []*model.Department, subEmployees []*model.Employee, audit *model.AuditLog, idempotencyKey string) error {
	prevVersion := dept.Version
	dept.Version++
	dept.UpdatedAt = time.Now().UTC()

	auditPK := fmt.Sprintf("AUDIT#%s", dept.ID)
	auditSK := fmt.Sprintf("LOG#%s#%s", audit.OccurredAt.Format(time.RFC3339Nano), audit.ID)
	parentGSI := dept.ParentID
	if parentGSI == "" {
		parentGSI = "ROOT"
	}

	var actions []types.TransactWriteItem

	actions = append(actions, types.TransactWriteItem{
		Update: &types.Update{
			TableName: aws.String(r.tableName),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPT#%s", dept.ID)},
				"SK": &types.AttributeValueMemberS{Value: "METADATA"},
			},
			UpdateExpression:    aws.String("SET parentId = :pId, #p = :path, version = :newVer, GSI1PK = :gpk, GSI1SK = :gsk, updatedAt = :uAt"),
			ConditionExpression: aws.String("version = :oldVer"),
			ExpressionAttributeNames: map[string]string{
				"#p": "path",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pId":    &types.AttributeValueMemberS{Value: dept.ParentID},
				":path":   &types.AttributeValueMemberS{Value: dept.Path},
				":newVer": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", dept.Version)},
				":oldVer": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", prevVersion)},
				":gpk":    &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPT_PARENT#%s", parentGSI)},
				":gsk":    &types.AttributeValueMemberS{Value: fmt.Sprintf("PATH#%s", dept.Path)},
				":uAt":    &types.AttributeValueMemberS{Value: dept.UpdatedAt.Format(time.RFC3339)},
			},
		},
	})
	if idempotencyKey != "" {
		actions = append(actions, idempotencyMessagePut(r.tableName, idempotencyKey, "Di chuyển phòng ban thành công"))
	}

	for _, sd := range subDepts {
		actions = append(actions, types.TransactWriteItem{
			Update: &types.Update{
				TableName: aws.String(r.tableName),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPT#%s", sd.ID)},
					"SK": &types.AttributeValueMemberS{Value: "METADATA"},
				},
				UpdateExpression: aws.String("SET #p = :path, GSI1SK = :gsk, updatedAt = :uAt"),
				ExpressionAttributeNames: map[string]string{
					"#p": "path",
				},
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":path": &types.AttributeValueMemberS{Value: sd.Path},
					":gsk":  &types.AttributeValueMemberS{Value: fmt.Sprintf("PATH#%s", sd.Path)},
					":uAt":  &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
				},
			},
		})
	}

	for _, se := range subEmployees {
		actions = append(actions, types.TransactWriteItem{
			Update: &types.Update{
				TableName: aws.String(r.tableName),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("EMP#%s", se.ID)},
					"SK": &types.AttributeValueMemberS{Value: "METADATA"},
				},
				UpdateExpression: aws.String("SET departmentPath = :dp, updatedAt = :uAt"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":dp":  &types.AttributeValueMemberS{Value: se.DepartmentPath},
					":uAt": &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
				},
			},
		})
	}

	auditItem, _ := attributevalue.MarshalMap(audit)
	auditItem["PK"] = &types.AttributeValueMemberS{Value: auditPK}
	auditItem["SK"] = &types.AttributeValueMemberS{Value: auditSK}
	auditItem["GSI1PK"] = &types.AttributeValueMemberS{Value: "AUDIT_ALL"}
	auditItem["GSI1SK"] = &types.AttributeValueMemberS{Value: auditSK}

	actions = append(actions, types.TransactWriteItem{
		Put: &types.Put{
			TableName: aws.String(r.tableName),
			Item:      auditItem,
		},
	})

	_, err := r.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: actions})
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailed") || strings.Contains(err.Error(), "TransactionCanceledException") {
			return ErrVersionConflict
		}
		return err
	}
	return nil
}

func (r *DynamoRepository) ArchiveDepartmentAtomic(ctx context.Context, dept *model.Department, audit *model.AuditLog, idempotencyKey string) error {
	prevVersion := dept.Version
	dept.Version++
	dept.Status = model.DeptStatusArchived
	dept.UpdatedAt = time.Now().UTC()

	auditPK := fmt.Sprintf("AUDIT#%s", dept.ID)
	auditSK := fmt.Sprintf("LOG#%s#%s", audit.OccurredAt.Format(time.RFC3339Nano), audit.ID)

	auditItem, _ := attributevalue.MarshalMap(audit)
	auditItem["PK"] = &types.AttributeValueMemberS{Value: auditPK}
	auditItem["SK"] = &types.AttributeValueMemberS{Value: auditSK}
	auditItem["GSI1PK"] = &types.AttributeValueMemberS{Value: "AUDIT_ALL"}
	auditItem["GSI1SK"] = &types.AttributeValueMemberS{Value: auditSK}

	actions := []types.TransactWriteItem{
		{
			Update: &types.Update{
				TableName: aws.String(r.tableName),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPT#%s", dept.ID)},
					"SK": &types.AttributeValueMemberS{Value: "METADATA"},
				},
				UpdateExpression:    aws.String("SET #s = :st, version = :newVer, updatedAt = :uAt"),
				ConditionExpression: aws.String("version = :oldVer AND #s = :active"),
				ExpressionAttributeNames: map[string]string{
					"#s": "status",
				},
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":st":     &types.AttributeValueMemberS{Value: string(model.DeptStatusArchived)},
					":active": &types.AttributeValueMemberS{Value: string(model.DeptStatusActive)},
					":newVer": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", dept.Version)},
					":oldVer": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", prevVersion)},
					":uAt":    &types.AttributeValueMemberS{Value: dept.UpdatedAt.Format(time.RFC3339)},
				},
			},
		},
		{
			Put: &types.Put{
				TableName: aws.String(r.tableName),
				Item:      auditItem,
			},
		},
	}
	if idempotencyKey != "" {
		actions = append(actions, idempotencyMessagePut(r.tableName, idempotencyKey, "Lưu trữ phòng ban thành công"))
	}

	_, err := r.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: actions})
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailed") || strings.Contains(err.Error(), "TransactionCanceledException") {
			return ErrVersionConflict
		}
		return err
	}
	return nil
}

func (r *DynamoRepository) PutEmployeeAtomic(ctx context.Context, emp *model.Employee, audit *model.AuditLog, idempotencyKey string) error {
	empPK := fmt.Sprintf("EMP#%s", emp.ID)
	codePK := fmt.Sprintf("EMP_CODE#%s", emp.Code)
	emailPK := fmt.Sprintf("EMP_EMAIL#%s", emp.Email)
	auditPK := fmt.Sprintf("AUDIT#%s", emp.ID)
	auditSK := fmt.Sprintf("LOG#%s#%s", audit.OccurredAt.Format(time.RFC3339Nano), audit.ID)

	empItem, err := attributevalue.MarshalMap(emp)
	if err != nil {
		return err
	}
	empItem["PK"] = &types.AttributeValueMemberS{Value: empPK}
	empItem["SK"] = &types.AttributeValueMemberS{Value: "METADATA"}
	empItem["GSI1PK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPT#%s", emp.DepartmentID)}
	empItem["GSI1SK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("STATUS_JOINED#%s#%s", emp.Status, emp.JoinedAt)}

	auditItem, _ := attributevalue.MarshalMap(audit)
	auditItem["PK"] = &types.AttributeValueMemberS{Value: auditPK}
	auditItem["SK"] = &types.AttributeValueMemberS{Value: auditSK}
	auditItem["GSI1PK"] = &types.AttributeValueMemberS{Value: "AUDIT_ALL"}
	auditItem["GSI1SK"] = &types.AttributeValueMemberS{Value: auditSK}

	actions := []types.TransactWriteItem{
		{
			Put: &types.Put{
				TableName:           aws.String(r.tableName),
				Item:                empItem,
				ConditionExpression: aws.String("attribute_not_exists(PK)"),
			},
		},
		{
			Put: &types.Put{
				TableName: aws.String(r.tableName),
				Item: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: codePK},
					"SK": &types.AttributeValueMemberS{Value: "METADATA"},
				},
				ConditionExpression: aws.String("attribute_not_exists(PK)"),
			},
		},
		{
			Put: &types.Put{
				TableName: aws.String(r.tableName),
				Item: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: emailPK},
					"SK": &types.AttributeValueMemberS{Value: "METADATA"},
				},
				ConditionExpression: aws.String("attribute_not_exists(PK)"),
			},
		},
		{
			Update: &types.Update{
				TableName: aws.String(r.tableName),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPT#%s", emp.DepartmentID)},
					"SK": &types.AttributeValueMemberS{Value: "METADATA"},
				},
				UpdateExpression:          aws.String("ADD employeeCount :inc"),
				ExpressionAttributeValues: map[string]types.AttributeValue{":inc": &types.AttributeValueMemberN{Value: "1"}},
			},
		},
		{
			Put: &types.Put{
				TableName: aws.String(r.tableName),
				Item:      auditItem,
			},
		},
	}
	if emp.Status == model.EmpStatusResigned {
		actions = append(actions[:3], actions[4:]...)
	}
	if idempotencyKey != "" {
		response, _ := json.Marshal(map[string]interface{}{"success": true, "data": emp})
		actions = append(actions, idempotencyPut(r.tableName, idempotencyKey, http.StatusCreated, response))
	}

	_, err = r.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: actions})
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailed") || strings.Contains(err.Error(), "TransactionCanceledException") {
			return ErrDuplicateCode
		}
		return err
	}
	return nil
}

func idempotencyPut(table, key string, status int, response []byte) types.TransactWriteItem {
	return types.TransactWriteItem{Put: &types.Put{
		TableName: aws.String(table),
		Item: map[string]types.AttributeValue{
			"PK":         &types.AttributeValueMemberS{Value: "IDEMPOTENCY#" + key},
			"SK":         &types.AttributeValueMemberS{Value: "METADATA"},
			"statusCode": &types.AttributeValueMemberN{Value: fmt.Sprint(status)},
			"response":   &types.AttributeValueMemberS{Value: string(response)},
			"expiresAt":  &types.AttributeValueMemberN{Value: fmt.Sprint(time.Now().Add(24 * time.Hour).Unix())},
		},
		ConditionExpression:       aws.String("attribute_not_exists(PK) OR expiresAt < :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":now": &types.AttributeValueMemberN{Value: fmt.Sprint(time.Now().Unix())}},
	}}
}

func idempotencyMessagePut(table, key, message string) types.TransactWriteItem {
	response, _ := json.Marshal(map[string]interface{}{"success": true, "data": map[string]string{"message": message}})
	return idempotencyPut(table, key, http.StatusOK, response)
}

func (r *DynamoRepository) GetEmployeeByID(ctx context.Context, id string) (*model.Employee, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("EMP#%s", id)},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, ErrNotFound
	}

	var emp model.Employee
	err = attributevalue.UnmarshalMap(out.Item, &emp)
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *DynamoRepository) UpdateEmployeeAtomic(ctx context.Context, employee *model.Employee, audit *model.AuditLog, avatarChanged bool) error {
	oldVersion := employee.Version
	employee.Version++
	employee.UpdatedAt = time.Now().UTC()
	item, err := attributevalue.MarshalMap(audit)
	if err != nil {
		return err
	}
	item["PK"] = &types.AttributeValueMemberS{Value: "AUDIT#" + employee.ID}
	item["SK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("LOG#%s#%s", audit.OccurredAt.Format(time.RFC3339Nano), audit.ID)}
	item["GSI1PK"] = &types.AttributeValueMemberS{Value: "AUDIT_ALL"}
	item["GSI1SK"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("LOG#%s#%s", audit.OccurredAt.Format(time.RFC3339Nano), audit.ID)}
	updateExpression := "SET fullName = :name, title = :title, joinedAt = :joined, GSI1SK = :gsk, version = :next, updatedAt = :updated"
	values := map[string]types.AttributeValue{
		":name": &types.AttributeValueMemberS{Value: employee.FullName}, ":title": &types.AttributeValueMemberS{Value: employee.Title},
		":joined": &types.AttributeValueMemberS{Value: employee.JoinedAt}, ":gsk": &types.AttributeValueMemberS{Value: fmt.Sprintf("STATUS_JOINED#%s#%s", employee.Status, employee.JoinedAt)},
		":old": &types.AttributeValueMemberN{Value: fmt.Sprint(oldVersion)}, ":next": &types.AttributeValueMemberN{Value: fmt.Sprint(employee.Version)}, ":updated": &types.AttributeValueMemberS{Value: employee.UpdatedAt.Format(time.RFC3339)},
	}
	if avatarChanged {
		if employee.Avatar == "" {
			updateExpression += " REMOVE avatar"
		} else {
			updateExpression += ", avatar = :avatar"
			values[":avatar"] = &types.AttributeValueMemberS{Value: employee.Avatar}
		}
	}
	_, err = r.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: []types.TransactWriteItem{
		{Update: &types.Update{
			TableName:                 aws.String(r.tableName),
			Key:                       map[string]types.AttributeValue{"PK": &types.AttributeValueMemberS{Value: "EMP#" + employee.ID}, "SK": &types.AttributeValueMemberS{Value: "METADATA"}},
			UpdateExpression:          aws.String(updateExpression),
			ConditionExpression:       aws.String("version = :old"),
			ExpressionAttributeValues: values,
		}},
		{Put: &types.Put{TableName: aws.String(r.tableName), Item: item}},
	}})
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailed") || strings.Contains(err.Error(), "TransactionCanceledException") {
			return ErrVersionConflict
		}
		return err
	}
	return nil
}

func (r *DynamoRepository) TransferEmployeeAtomic(ctx context.Context, emp *model.Employee, oldDeptID string, clearOldDeptManager bool, audit *model.AuditLog, idempotencyKey string) error {
	prevVersion := emp.Version
	emp.Version++
	emp.UpdatedAt = time.Now().UTC()

	auditPK := fmt.Sprintf("AUDIT#%s", emp.ID)
	auditSK := fmt.Sprintf("LOG#%s#%s", audit.OccurredAt.Format(time.RFC3339Nano), audit.ID)
	oldDepartmentUpdateExpression := "ADD employeeCount :dec"
	if clearOldDeptManager {
		// A DynamoDB transaction may touch a given item only once. Combine the
		// headcount decrement and manager removal into this single department update.
		oldDepartmentUpdateExpression = "REMOVE managerId, managerName ADD employeeCount :dec"
	}

	actions := []types.TransactWriteItem{
		{
			Update: &types.Update{
				TableName: aws.String(r.tableName),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("EMP#%s", emp.ID)},
					"SK": &types.AttributeValueMemberS{Value: "METADATA"},
				},
				UpdateExpression:    aws.String("SET departmentId = :dId, departmentName = :dName, departmentPath = :dPath, GSI1PK = :gpk, version = :newVer, updatedAt = :uAt"),
				ConditionExpression: aws.String("version = :oldVer"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":dId":    &types.AttributeValueMemberS{Value: emp.DepartmentID},
					":dName":  &types.AttributeValueMemberS{Value: emp.DepartmentName},
					":dPath":  &types.AttributeValueMemberS{Value: emp.DepartmentPath},
					":gpk":    &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPT#%s", emp.DepartmentID)},
					":newVer": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", emp.Version)},
					":oldVer": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", prevVersion)},
					":uAt":    &types.AttributeValueMemberS{Value: emp.UpdatedAt.Format(time.RFC3339)},
				},
			},
		},
		{
			Update: &types.Update{
				TableName: aws.String(r.tableName),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPT#%s", oldDeptID)},
					"SK": &types.AttributeValueMemberS{Value: "METADATA"},
				},
				UpdateExpression: aws.String(oldDepartmentUpdateExpression),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":dec": &types.AttributeValueMemberN{Value: "-1"},
				},
			},
		},
		{
			Update: &types.Update{
				TableName: aws.String(r.tableName),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPT#%s", emp.DepartmentID)},
					"SK": &types.AttributeValueMemberS{Value: "METADATA"},
				},
				UpdateExpression: aws.String("ADD employeeCount :inc"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":inc": &types.AttributeValueMemberN{Value: "1"},
				},
			},
		},
	}

	auditItem, _ := attributevalue.MarshalMap(audit)
	auditItem["PK"] = &types.AttributeValueMemberS{Value: auditPK}
	auditItem["SK"] = &types.AttributeValueMemberS{Value: auditSK}
	auditItem["GSI1PK"] = &types.AttributeValueMemberS{Value: "AUDIT_ALL"}
	auditItem["GSI1SK"] = &types.AttributeValueMemberS{Value: auditSK}

	actions = append(actions, types.TransactWriteItem{
		Put: &types.Put{
			TableName: aws.String(r.tableName),
			Item:      auditItem,
		},
	})
	if idempotencyKey != "" {
		actions = append(actions, idempotencyMessagePut(r.tableName, idempotencyKey, "Chuyển phòng ban cho nhân sự thành công"))
	}

	_, err := r.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: actions})
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailed") || strings.Contains(err.Error(), "TransactionCanceledException") {
			return ErrVersionConflict
		}
		return err
	}
	return nil
}

func (r *DynamoRepository) ResignEmployeeAtomic(ctx context.Context, emp *model.Employee, audit *model.AuditLog, idempotencyKey string) error {
	prevVersion := emp.Version
	emp.Version++
	emp.Status = model.EmpStatusResigned
	emp.UpdatedAt = time.Now().UTC()

	auditPK := fmt.Sprintf("AUDIT#%s", emp.ID)
	auditSK := fmt.Sprintf("LOG#%s#%s", audit.OccurredAt.Format(time.RFC3339Nano), audit.ID)

	auditItem, _ := attributevalue.MarshalMap(audit)
	auditItem["PK"] = &types.AttributeValueMemberS{Value: auditPK}
	auditItem["SK"] = &types.AttributeValueMemberS{Value: auditSK}
	auditItem["GSI1PK"] = &types.AttributeValueMemberS{Value: "AUDIT_ALL"}
	auditItem["GSI1SK"] = &types.AttributeValueMemberS{Value: auditSK}

	actions := []types.TransactWriteItem{
		{
			Update: &types.Update{
				TableName: aws.String(r.tableName),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("EMP#%s", emp.ID)},
					"SK": &types.AttributeValueMemberS{Value: "METADATA"},
				},
				UpdateExpression:    aws.String("SET #s = :st, resignedAt = :rAt, resignReason = :rR, GSI1SK = :gsk, version = :newVer, updatedAt = :uAt"),
				ConditionExpression: aws.String("version = :oldVer AND #s = :active"),
				ExpressionAttributeNames: map[string]string{
					"#s": "status",
				},
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":st":     &types.AttributeValueMemberS{Value: string(model.EmpStatusResigned)},
					":active": &types.AttributeValueMemberS{Value: string(model.EmpStatusActive)},
					":rAt":    &types.AttributeValueMemberS{Value: emp.ResignedAt},
					":rR":     &types.AttributeValueMemberS{Value: emp.ResignReason},
					":gsk":    &types.AttributeValueMemberS{Value: fmt.Sprintf("STATUS_JOINED#%s#%s", model.EmpStatusResigned, emp.JoinedAt)},
					":newVer": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", emp.Version)},
					":oldVer": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", prevVersion)},
					":uAt":    &types.AttributeValueMemberS{Value: emp.UpdatedAt.Format(time.RFC3339)},
				},
			},
		},
		{
			Update: &types.Update{
				TableName: aws.String(r.tableName),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPT#%s", emp.DepartmentID)},
					"SK": &types.AttributeValueMemberS{Value: "METADATA"},
				},
				UpdateExpression: aws.String("ADD employeeCount :dec"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":dec": &types.AttributeValueMemberN{Value: "-1"},
				},
			},
		},
		{
			Put: &types.Put{
				TableName: aws.String(r.tableName),
				Item:      auditItem,
			},
		},
	}
	if idempotencyKey != "" {
		actions = append(actions, idempotencyMessagePut(r.tableName, idempotencyKey, "Cập nhật thôi việc nhân sự thành công"))
	}

	_, err := r.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: actions})
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailed") || strings.Contains(err.Error(), "TransactionCanceledException") {
			return ErrVersionConflict
		}
		return err
	}
	return nil
}

func (r *DynamoRepository) ListAuditLogs(ctx context.Context, targetID string) ([]*model.AuditLog, error) {
	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("AUDIT#%s", targetID)},
			":sk": &types.AttributeValueMemberS{Value: "LOG#"},
		},
		ScanIndexForward: aws.Bool(false),
	})
	if err != nil {
		return nil, err
	}

	var logs []*model.AuditLog
	for _, it := range out.Items {
		var log model.AuditLog
		_ = attributevalue.UnmarshalMap(it, &log)
		logs = append(logs, &log)
	}
	return logs, nil
}

func (r *DynamoRepository) ListAllAuditLogs(ctx context.Context) ([]*model.AuditLog, error) {
	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :gpk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gpk": &types.AttributeValueMemberS{Value: "AUDIT_ALL"},
		},
		ScanIndexForward: aws.Bool(false),
	})
	if err != nil {
		return nil, err
	}

	var logs []*model.AuditLog
	for _, it := range out.Items {
		var log model.AuditLog
		_ = attributevalue.UnmarshalMap(it, &log)
		logs = append(logs, &log)
	}
	return logs, nil
}
