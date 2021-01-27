// +build linux darwin

/*
Package jmx is a library to get metrics through JMX. It requires additional
setup. Read https://github.com/newrelic/infra-integrations-sdk#jmx-support for
instructions. */
package jmx

const (
	defaultNrjmxExec = "/usr/bin/nrjmx" // defaultNrjmxExec default nrjmx tool executable path
)
