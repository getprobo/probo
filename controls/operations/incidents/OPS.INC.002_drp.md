---
id: "OPS.INC.002"
category: "operations/incidents"
revision-version: 1
revision-date: "2024-01-14"
estimate-time: "6h"
necessity: "mandatory"
frameworks:
  - name: "soc2"
    sections: ["CC2.2"]
---

# Disaster Recovery Plan (DRP)

## Purpose

Disasters might seem far-fetched, but data loss, service outages, and
misconfigurations can happen. Whether it's a simple human error or a regional
outage, a DRP helps your team recover fast and maintain customer trust.

## Implementation

Here is a basic template, to use as a good starting point:

[DRP Template](https://www.notion.so/DRP-Template-13e1cc0bd5bc800d9db2f77d3c884521?pvs=21)

<aside>
💡

If it is easier to manage, you can have several DRP: one for your database, one
for your network, one for your infra, etc… that way, when testing, you can test
smaller part of it and it adapts better to potential real life events.

</aside>

The idea is to be prepared for the unexpected. You don’t need to overthink
things, it’s more to know what to do when you will be in the rush of a regional
outage. For that, your DRP needs to cover:

1. **Scope & timings**: clear scope and the expected time to do it
2. **Step-by-step:** guide on how to proceed
3. **Contacts**: list of internal and external contacts crucial during a
   disaster.
4. **Testing and Review Log**: A record of tests performed to validate the DRP’s
   effectiveness and any improvements made.
