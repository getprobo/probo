---
id: "OPS.INC.001"
category: "operations/incidents"
revision-version: 1
revision-date: "2024-01-14"
estimate-time: "30m"
necessity: "mandatory"
frameworks:
  - name: "soc2"
    sections: ["CC2.2"]
---

# Build an incident response process

## Purpose

An incident response plan is crucial for quickly identifying, containing, and
resolving incidents, minimizing potential disruptions. It ensures you’re
prepared effectively and that you keep your operations running smoothly, even in
the face of unexpected threats.

## Implementation

### Build an Incident Response Plan and share it with your team.

Template:

[Incident response plan](templates/incident_response_plan.md)

### Create a slack channel

Create a slack channel to allow your team to raise incident and be kept aware of
their resolution

1. Create a channel (eg `#incident`)
2. Pin the **Incident Response Plan** in the channel for quick access.
3. Establish guidelines for using the channel, including:
   1. How to report incidents (e.g., “Post a message with a brief summary and
      relevant context”).
   2. Expected response times for acknowledgment and updates.
4. Assign roles for incident management, such as:
   1. **On-duty engineer**: Monitors and triages notifications.
   2. **Response team**: Engineers responsible for handling and resolving
      incidents.
5. Set up reminders or periodic messages in Slack to encourage proactive
   monitoring and review of unresolved issues.
6. Integrate relevant notifications, if applicable, into the Slack channel for
   visibility (e.g., system alerts or logs).

## Evidence

- Share your incident response plan
- screenshot of your #incident slack channel being used (with postmortem
  visible)
