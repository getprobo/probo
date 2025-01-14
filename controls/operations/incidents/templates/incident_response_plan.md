### **1. Triage & validate**

The on-duty engineer receives the notification, he/she analyze logs, system
behaviors, and context related to the indication versus known behaviors.

He/She needs to determine whether the notification is legit or not:

1. If it is legit ⇒ it is promoted to an ongoing incident. In the thread, CTO is
   notified and a message is posted with:

   - A brief summary of the incident
   - Any immediate risks or concerns
   - Preliminary findings and scope (if known)
   - Ping for the relevant engineer (the response team).

   ⇒ the response team commences work.

2. If it is not (ie a false alarm), a message is posted in the thread with the
   reason for dismissal. The on-duty engineer needs to:
   - Determine and address the root cause for the false alarm. He/she needs to
     remediate any system or process issues that led to it.
   - End the incident response process

### **2. Handling & Resolution**

Resolution occurs in 3 steps detailed below. All decisions or actions taken must
be documented.

1. **Contain:**

If possible, isolate system to prevent the incident from spreading or causing
additional damage

1. **Eradicate:**

Identify and remove the root cause of the incident - roll back is a strong
option.

1. **Recover:**

Resume normal business operations by restoring system functionality. Keep strong
monitoring to catch any signs of repetition.

At the end of this process, an update is posted on the thread to keep all team
informed. If the complete process takes longer than 30min, a quick update will
be given in the thread every 30min so stakeholders are aware of the current
state.

### **3. Postmortem**

Once the incident is resolved, the on-duty engineer document the findings,
decisions and actions taken during the event in a document to which he/she joins
recommendations. This document must be shared with stakeholders.

The postmortem document must answer:

- What was the incident? What happened?
- Which decisions or actions were taken to respond?
- What could have prevented it from happening?
- What can be improved in terms of security, processes, communication, tools,
  etc.?

The CTO must validate and ensure that recommended improvements are implemented.
