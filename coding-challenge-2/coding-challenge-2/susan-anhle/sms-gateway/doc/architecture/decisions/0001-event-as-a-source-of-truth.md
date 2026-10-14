# 1. Audit record as a source of truth

Date: 2026-04-14

## Status

Accepted

## Context

We are building an SMS gateway. We need to choose should we store the audit record or should we store the current state of the SMS and update it with each state transition

## Decision

We will use audit records as the source of truth. We can easily reconstruct current state from the audit records, and it also provides a complete history of state transitions for each message.
As audit records can also support data analysis relevant requirements in the future

## Consequences

Enables full auditability and traceability of all changes. 
Supports temporal queries and easier debugging of historical issues.
Facilitates integration with other systems. 

Increases storage requirements and complexity in reconstructing state. 
Requires careful design of event schemas and versioning strategies.
New developers might not be familiar with this pattern. The team need to be aware and have a coaching strategy to keep the code base consistent and avoid confusion. 

