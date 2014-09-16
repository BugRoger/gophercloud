package containers

import (
	"net/http"

	"github.com/racker/perigee"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/utils"
	"github.com/rackspace/gophercloud/pagination"
)

// ListResult is a *http.Response that is returned from a call to the List function.
type ListResult struct {
	pagination.MarkerPageBase
}

// IsEmpty returns true if a ListResult contains no container names.
func (r ListResult) IsEmpty() (bool, error) {
	names, err := ExtractNames(r)
	if err != nil {
		return true, err
	}
	return len(names) == 0, nil
}

// LastMarker returns the last container name in a ListResult.
func (r ListResult) LastMarker() (string, error) {
	names, err := ExtractNames(r)
	if err != nil {
		return "", err
	}
	if len(names) == 0 {
		return "", nil
	}
	return names[len(names)-1], nil
}

// GetResult is a *http.Response that is returned from a call to the Get function.
type GetResult *http.Response

// List is a function that retrieves all objects in a container. It also returns the details
// for the account. To extract just the container information or names, pass the ListResult
// response to the ExtractInfo or ExtractNames function, respectively.
func List(c *gophercloud.ServiceClient, opts ListOpts) pagination.Pager {
	var headers map[string]string

	query := utils.BuildQuery(opts.Params)

	if !opts.Full {
		headers = map[string]string{"Content-Type": "text/plain"}
	}

	createPage := func(r pagination.LastHTTPResponse) pagination.Page {
		p := ListResult{pagination.MarkerPageBase{LastHTTPResponse: r}}
		p.MarkerPageBase.Owner = p
		return p
	}

	url := getAccountURL(c) + query
	pager := pagination.NewPager(c, url, createPage)
	pager.Headers = headers
	return pager
}

// Create is a function that creates a new container.
func Create(c *gophercloud.ServiceClient, opts CreateOpts) (Container, error) {
	var ci Container

	h := c.Provider.AuthenticatedHeaders()

	for k, v := range opts.Headers {
		h[k] = v
	}

	for k, v := range opts.Metadata {
		h["X-Container-Meta-"+k] = v
	}

	url := getContainerURL(c, opts.Name)
	_, err := perigee.Request("PUT", url, perigee.Options{
		MoreHeaders: h,
		OkCodes:     []int{201, 204},
	})
	if err == nil {
		ci = Container{
			"name": opts.Name,
		}
	}
	return ci, err
}

// Delete is a function that deletes a container.
func Delete(c *gophercloud.ServiceClient, opts DeleteOpts) error {
	h := c.Provider.AuthenticatedHeaders()

	query := utils.BuildQuery(opts.Params)

	url := getContainerURL(c, opts.Name) + query
	_, err := perigee.Request("DELETE", url, perigee.Options{
		MoreHeaders: h,
		OkCodes:     []int{204},
	})
	return err
}

// Update is a function that creates, updates, or deletes a container's metadata.
func Update(c *gophercloud.ServiceClient, opts UpdateOpts) error {
	h := c.Provider.AuthenticatedHeaders()

	for k, v := range opts.Headers {
		h[k] = v
	}

	for k, v := range opts.Metadata {
		h["X-Container-Meta-"+k] = v
	}

	url := getContainerURL(c, opts.Name)
	_, err := perigee.Request("POST", url, perigee.Options{
		MoreHeaders: h,
		OkCodes:     []int{204},
	})
	return err
}

// Get is a function that retrieves the metadata of a container. To extract just the custom
// metadata, pass the GetResult response to the ExtractMetadata function.
func Get(c *gophercloud.ServiceClient, opts GetOpts) (GetResult, error) {
	h := c.Provider.AuthenticatedHeaders()

	for k, v := range opts.Metadata {
		h["X-Container-Meta-"+k] = v
	}

	url := getContainerURL(c, opts.Name)
	resp, err := perigee.Request("HEAD", url, perigee.Options{
		MoreHeaders: h,
		OkCodes:     []int{204},
	})
	return &resp.HttpResponse, err
}
