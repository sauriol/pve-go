package main


// TODO:
//  - Add the rest of `/access`
//  - Add custom structs for domains, roles, and users
//  - Implement passing said struct to create/edit

import (
    "net/url"
    "github.com/mitchellh/mapstructure"
)

type Group struct {
    Name    string
    Comment string
    Members []string
}

type Domain struct {
    Comment string
    Digest  string
    Plugin  string
    Type    string
}

// Returns the valid subdirectories for `/access`. For example, because `roles`
// is returned, `/access/roles` is a valid API path
func (proxmox Proxmox) Access() ([]string, error) {
    data, err := proxmox.Get("/access")
    if err != nil {
        return nil, err
    }

    var subdirs []string
    dataArr := data.([]interface{})
    for _, element := range dataArr {
        elementMap := element.(map[string] interface{})
        subdirs = append(subdirs, elementMap["subdir"].(string))
    }

    return subdirs, nil
}

func (proxmox Proxmox) GetAccessDomains() ([]*Domain, error) {
    var domains []*Domain
    data, err := proxmox.Get("/access/domains")
    if err != nil {
        return nil, err
    }
    dataArr := data.([]interface{})

    for _, element := range dataArr {
        elementMap := element.(map[string] interface{})
        domain, err := proxmox.GetAccessDomain(elementMap["realm"].(string))
        if err != nil {
            return nil, err
        }
        domains = append(domains, domain)
    }
    return domains, nil
}

// Untested
// TODO:
//  - Pass in options as Domain struct instead of form
func (proxmox Proxmox) AddAccessDomain(form url.Values) (map[string] interface{}, error) {
    data, err := proxmox.PostForm("/access/domains", form)
    if err != nil {
        return nil, err
    }
    dataMap := data.(map[string]interface{})
    return dataMap, err
}

func (proxmox Proxmox) GetAccessDomain(name string) (*Domain, error) {
    var domain Domain
    data, err := proxmox.Get("/access/domains/" + name)
    if err != nil {
        return nil, err
    }
    data = data.(map[string]interface{})

    err = mapstructure.Decode(data, &domain)
    if err != nil {
        return nil, err
    }
    return &domain, nil
}

// Untested
// TODO:
//  - Pass in Domain struct instead of form
func (proxmox Proxmox) EditAccessDomain (domain string,
    form url.Values) (map[string] interface{}, error) {
    data, err := proxmox.PostForm("/access/domains/" + domain, form)
    if err != nil {
        return nil, err
    }
    dataMap := data.(map[string]interface{})
    return dataMap, nil
}

// Untested
func (proxmox Proxmox) DeleteAccessDomain(domain string) error {
    _, err := proxmox.Delete("/access/domains/" + domain)
    if err != nil {
        return err
    }
    return nil
}

func (proxmox Proxmox) GetAccessGroups() ([]*Group, error) {
    data, err := proxmox.Get("/access/groups")
    if err != nil {
        return nil, err
    }
    dataMap := data.([]interface{})

    var groups []*Group
    for _, element := range dataMap {
        elementMap := element.(map[string] interface{})
        group, err := proxmox.GetAccessGroup(elementMap["groupid"].(string))
        if err != nil {
            return nil, err
        }
        groups = append(groups, group)
    }
    return groups, nil
}

// Untested
func (proxmox Proxmox) AddAccessGroup(form url.Values) (map[string] interface{}, error) {
    data, err := proxmox.PostForm("/access/groups", form)
    if err != nil {
        return nil, err
    }
    dataMap := data.(map[string]interface{})
    return dataMap, err
}

func (proxmox Proxmox) GetAccessGroup(name string) (*Group, error) {
    var group Group

    data, err := proxmox.Get("/access/groups/" + name)
    if err != nil {
        return nil, err
    }
    data = data.(map[string] interface{})

    err = mapstructure.Decode(data, &group)
    if err != nil {
        return nil, err
    }
    group.Name = name

    return &group, nil
}

// Untested
func (proxmox Proxmox) EditAccessGroup (name string, form url.Values) (map[string] interface{}, error) {
    data, err := proxmox.PostForm("/access/groups/" + name, form)
    if err != nil {
        return nil, err
    }
    dataMap := data.(map[string]interface{})
    return dataMap, nil
}

// Untested
func (proxmox Proxmox) DeleteAccessGroup(name string) error {
    _, err := proxmox.Delete("/access/domains/" + name)
    if err != nil {
        return err
    }
    return nil
}
