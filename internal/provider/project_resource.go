package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ProjectResource satisfies the resource.Resource interface.
var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithConfigure = &ProjectResource{}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

// ProjectResource implements the zicon_project resource.
type ProjectResource struct {
	client *Client
}

// ProjectResourceModel maps the zicon_project schema to Go types.
type ProjectResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Category    types.String `tfsdk:"category"`
	Description types.String `tfsdk:"description"`
	APIKey      types.String `tfsdk:"api_key"`
}

func (r *ProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a ZiCON Cloud project. There is no update endpoint, so any change to name, category, or description destroys and recreates the project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier of the project, assigned by the ZiCON Cloud API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"category": schema.StringAttribute{
				Optional:    true,
				Description: "Category of the project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"api_key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "API key issued for the project, used to authenticate subsequent requests to it.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.CreateProject(ctx, plan.Name.ValueString(), plan.Category.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Creating ZiCON Project", err.Error())
		return
	}

	plan.ID = types.StringValue(project.ID)
	plan.APIKey = types.StringValue(project.APIKey)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.ListProjects(ctx, state.APIKey.ValueString())
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			// The API key has been revoked or the project was deleted
			// out-of-band: drop it from state so Terraform recreates it.
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading ZiCON Project", err.Error())
		return
	}

	state.Name = types.StringValue(project.Name)
	state.Category = types.StringValue(project.Category)
	state.Description = types.StringValue(project.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All mutable attributes use RequiresReplace, so Terraform never calls
	// Update in practice; this exists only to satisfy resource.Resource.
	var plan ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteProject(ctx, state.APIKey.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting ZiCON Project", err.Error())
		return
	}
}
