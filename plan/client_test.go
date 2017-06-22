package plan

import (
	"testing"

	assert "github.com/stretchr/testify/require"
	stripe "github.com/stripe/stripe-go"
	_ "github.com/stripe/stripe-go/testing"
)

func TestPlanDel(t *testing.T) {
	plan, err := Del("gold")
	assert.Nil(t, err)
	assert.NotNil(t, plan)
}

func TestPlanGet(t *testing.T) {
	plan, err := Get("gold", nil)
	assert.Nil(t, err)
	assert.NotNil(t, plan)
}

func TestPlanList(t *testing.T) {
	i := List(&stripe.PlanListParams{})

	// Verify that we can get at least one plan
	assert.True(t, i.Next())
	assert.Nil(t, i.Err())
	assert.NotNil(t, i.Plan())
}

func TestPlanNew(t *testing.T) {
	plan, err := New(&stripe.PlanParams{
		Name: "Test Plan",
	})
	assert.Nil(t, err)
	assert.NotNil(t, plan)
}

func TestPlanUpdate(t *testing.T) {
	updatedPlan := &stripe.PlanParams{
		Name: "Updated Name",
	}
	plan, err := Update("gold", updatedPlan)
	assert.Nil(t, err)
	assert.NotNil(t, plan)
}
