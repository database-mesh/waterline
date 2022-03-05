<img src="static/waterline-logo-handdraw.png" alt="Waterline Logo" width="350" length="350"/>

[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://golang.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/database-mesh/waterline)](https://goreportcard.com/report/github.com/database-mesh/waterline)

> This readme and related documents are now WIP. 

# waterline

Waterline is a Database Mesh project. It provides QoS for SQL traffic in a cloud native way.

## Features

Different applications in production clusters are always applied with different priorities. For better SLA purpose, we treat application as different QoS class. At present, Kubernetes provides CPU and Memory QoS, and community has contributed some design for network QoS, such as ingress and egress traffic bandwidth. Waterline could provider a protocol-specific network QoS solution with the help of Traffic Control and eBPF.

## Community & Support

:link: [GitHub Issues](https://github.com/database-mesh/waterline/issues). Best for: larger systemic questions/bug reports or anything development related.

:link: [Slack channel](https://join.slack.com/t/databasemesh/shared_invite/zt-12hlythpe-C4rrS1WZ2ZkEd3zn84SqeQ). Best for: instant communications and online meetings, sharing your applications.
