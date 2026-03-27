## **System Overview**

You are working in a large-scale distributed system for a Cinema Ticket Booking Platform.

The system is already in production and serves multiple countries:

* Vietnam  
* Thailand  
* Singapore

The business is expanding to a new market:

* Philippines

The system consists of multiple teams, including:

* Customer Domain  
* Booking Domain  
* ERP System  
* Accounting & Operations
* Data & Analytics  
  

## **Your Responsibility**

Your team owns the Customer Domain, specifically: Membership Registration

A key requirement of membership registration is: Customers must receive OTP SMS verification during registration.

Assumption:

* OTP generation already exists (out of scope)  
* Your responsibility is SMS sending capability

## **Expanded Usage Context**

After introducing SMS capability, it becomes a shared platform capability used by multiple domains:

* Customer Domain → OTP, membership notifications  
* Booking Domain → ticket confirmation, reminders  
* ERP System → operational notifications  
* Accounting → cost tracking  
* Operations → provider optimization  
* Data Team → analytics & prediction 

## **SMS Business Context**

SMS delivery is not centralized. Instead:

* Each country has its own set of SMS providers, negotiated by local operations teams.  
* The system currently integrates (or may integrate) with providers such as:  
  * Infobip  
  * AWS SNS  
  * Vonage  
  * Twilio  
  * Telnyx  
  * MessageBird  
  * Sinch  
  * (and potentially others in the future)

### **Routing Complexity**

SMS routing depends on multiple factors:

1. Country  
2. Mobile Network Operator (MNO / Carrier) of the recipient  
3. Provider pricing and availability

Each provider:

* Has different pricing per carrier  
* May be blocked by certain carriers  
* Has changing pricing over time

### Example (Illustrative Only)

| Provider | Carrier | Price | Notes |
| ----- | ----- | ----- | ----- |
| Twilio | Mobifone | 0.02 | OK |
| Twilio | Viettel | 0.015 | OK |
| Twilio | Vinaphone | — | Blocked |
| Vonage | Mobifone | 0.018 | OK |
| Vonage | Viettel | 0.03 | OK |
| Vonage | Vinaphone | 0.012 | OK |

_These numbers are examples only_

### **Important Characteristics**

* Pricing changes over time  
* Routing decisions must consider:  
  * Cost optimization  
  * Carrier compatibility  
  * Operational constraints

## **SMS Sending Flow**

When an SMS is sent, it goes through the following states:

```
New 
→ Send-to-provider (cost estimate available) 
→ Queue 
→ Send-to-carrier (MNO)
→ Send-success (actual cost available)
```

**Alternative flows:**

```
Queue → Carrier-rejected → Cancel
```

### **Important Notes**

* All responses from providers are asynchronous  
* Providers will notify your system via webhooks or message callbacks  
* You do NOT need to implement real provider integrations  
  * Logging is sufficient to simulate sending/receiving

---

## **🎯 Epic: SMS Sending Capability for Multi-Domain Usage**

Design and implement a **reusable SMS capability** that:

* Supports Customer Membership OTP flow  
* Can be used by other domains (Booking, ERP, etc.)  
* Handles provider selection based on business constraints  
* Manages SMS lifecycle  
* Tracks cost and delivery outcomes

### **🧩 User Story 1 – Membership OTP SMS** 

**As a** Customer  
**I want** to receive an OTP via SMS during membership registration  
**So that** I can verify my identity

#### **Acceptance Criteria**

##### **Functional Behavior**

* When a membership registration is initiated  
  → An SMS must be sent containing OTP  
* The system must accept:  
  * messageId  
  * country  
  * phoneNumber  
  * message

The system must determine provider based on **country \+ carrier** as follows:

##### **Vietnam**

* Viettel → Twilio  
* Mobifone → Vonage  
* Vinaphone → Vonage

##### **Thailand**

* AIS → Infobip  
* DTAC → AWS SNS

##### **Singapore**

* Singtel → Twilio  
* StarHub → Telnyx

#### **Expected Behavior**

* Carrier must be derived from phone number (simulation allowed)  
* Exactly **one provider must be selected**  
* SMS must follow lifecycle:

```
New → Send-to-provider → Queue → Send-to-carrier → Send-success
```

#### **Cost Handling**

* Estimated cost must be available at **Send-to-provider**  
* Actual cost must be available at **Send-success**

### **🧩 User Story 2 – Cross-Domain SMS Usage**

**As a** Booking / ERP system  
**I want** to send SMS notifications  
**So that** I can communicate with customers or operations

#### **Acceptance Criteria**

* SMS capability must be usable by:  
  * Customer Domain  
  * Booking Domain  
  * ERP System  
* All domains must:  
  * Be able to trigger SMS sending  
  * Receive consistent behavior (routing, lifecycle, cost tracking)  
* SMS behavior must remain consistent regardless of caller, without requiring separate implementations per domain

### **🧩 User Story 3 – New Market and Updated Provider Agreements**

**As an** Operations team  
 **I want** SMS to be delivered using appropriate providers  
 **So** that delivery is reliable and cost-effective

Following new commercial agreements with SMS providers, routing rules have been updated in both new and existing markets.

#### **Acceptance Criteria**

##### **Philippines (New Market)**

* When sending SMS to Philippines:  
  * Globe → MessageBird  
  * Smart → Sinch  
  * DITO → MessageBird

##### **Vietnam (Updated Routing Rules)**

The routing rules in Vietnam have been updated based on new provider agreements.

* Viettel → Vonage  
* Mobifone → Infobip  
* Vinaphone → Twilio

##### **General Behavior**

* System must:  
  * Determine carrier from phone number  
  * Select exactly one provider  
* Provider selection must reflect current business rules 

### **🧩 User Story 4 – SMS Lifecycle Management**

**As a** system  
**I want** to manage SMS states clearly  
**So that** delivery status is traceable and reliable

#### **Acceptance Criteria**

##### **States**

The system must support:

```
New  
Send-to-provider  
Queue  
Send-to-carrier  
Send-success  
Send-failed  
Carrier-rejected  
```

##### **Transitions**

```
New → Send-to-provider  
Send-to-provider → Queue  
Queue → Send-to-carrier  
Send-to-carrier → Send-success  

Queue → Carrier-rejected → Send-to-provider  

Send-to-carrier → Send-failed → Send-to-provider  
```

##### **Behavior**

* All transitions must:  
  * Be valid and controlled  
  * Be traceable  
  * Prevent invalid transitions

#### **State Clarification**

* **Carrier-rejected**  
   The SMS was accepted by the provider but rejected by the downstream carrier (MNO).  
* **Send-failed**  
   The SMS failed during provider processing or delivery attempt before successful completion.

### **🧩 User Story 5 – Cost Tracking & Observability**

**As an** Accounting / Operations team  
**I want** visibility into SMS usage and cost  
**So that** I can monitor and optimize

#### **Acceptance Criteria**

* System must track:  
  * Estimated cost  
  * Actual cost  
* It must be possible to determine:  
  * Total cost per provider  
  * Total cost per country  
  * SMS volume per provider  
  * Success / failure rates

### **🧩 User Story 6 – Asynchronous Handling**

**As a** system  
**I want** to handle provider callbacks asynchronously  
**So that** SMS state is updated correctly

#### **Acceptance Criteria**

* System must support simulated callbacks:  
  * messageId  
  * provider  
  * newState  
  * actualCost  
* Callback must:  
  * Update SMS state  
  * Maintain correct lifecycle

---

## **Technical Expectations**

You are required to design and implement the SMS module with the following goals:

### **1\. Provider Integration Capability**

The system should support interaction with multiple SMS providers.

* Different providers may have:  
  * Different behaviors  
  * Different cost structures  
  * Different delivery constraints  
* The system should allow:  
  * Introducing additional providers  
  * Adjusting provider usage behavior over time

👉 The implementation should not assume a fixed or limited set of providers.

### **2\. Routing Behavior Adaptability**

SMS delivery depends on business rules involving:

* Country  
* Carrier (MNO)  
* Operational constraints

Over time, these rules may:

* Change for existing countries  
* Be introduced for new countries  
* Differ across business scenarios

The system should:

* Determine a provider based on the input context  
* Reflect current business rules accurately  
* Be able to accommodate changes in routing behavior

👉 The routing logic is expected to evolve as the business expands.

### **3\. State Management Clarity**

The SMS lifecycle includes multiple states and transitions.

The system should:

* Represent lifecycle states explicitly  
* Enforce valid transitions  
* Maintain traceability of state changes

As the system evolves:

* Additional states or transitions may be introduced  
* Existing flows may be adjusted

👉 State handling should remain clear and maintainable as complexity grows.

### **4\. Asynchronous Event Handling**

SMS delivery involves asynchronous interactions:

* Providers respond via callbacks or events  
* State updates occur after initial request

The system should:

* Handle asynchronous updates correctly  
* Ensure consistency between events and current state  
* Allow simulation of external callbacks

### **5\. Cross-Domain Reusability**

The SMS capability is used by multiple domains:

* Customer  
* Booking  
* ERP

The system should:

* Not assume a single domain usage  
* Provide consistent behavior regardless of caller  
* Be usable as a shared capability

👉 The design should support reuse without requiring domain-specific changes.

### **6\. Observability & Traceability**

The system should provide sufficient visibility into SMS processing:

* Provider selection  
* State transitions  
* Cost information

It should be possible to:

* Track delivery outcomes  
* Understand how a message flows through the system  
* Analyze usage across providers and countries

### **7\. Implementation Simplicity**

For the purpose of this challenge:

* External integrations can be simulated:  
  * Sending SMS → log output  
  * Provider callback → simulated input  
* Infrastructure is not required:  
  * No external APIs  
  * No real message queues  
  * No database persistence  
* Use:  
  * In-memory storage  
  * CLI commands or simple triggers

### **8\. Code Quality & Structure**

The implementation should demonstrate:

* Clear separation of concerns  
* Readable and maintainable code  
* Logical organization of components

Avoid:

* Mixing responsibilities  
* Embedding multiple concerns in a single component

## **🎯 Design Signal (Implicit)**

The system is expected to operate in an environment where:

* Providers may change  
* Routing behavior may evolve  
* Lifecycle complexity may increase  
* More domains may adopt the capability

👉 A strong solution should remain stable under such changes without requiring significant restructuring.

## **📦 Expected Deliverables**

Candidates are expected to submit:

* Source code of the solution  
* A short README including:  
  * How to run the application  
  * Example commands or scenarios  
* Any assumptions made during implementation

Optional (but recommended):

* Brief design explanation (high-level structure, key decisions) 

## **🚫 Out of Scope**

The following items are **explicitly out of scope** for this challenge and **must not be considered** in the design or implementation:

* Cost optimization  
* Carrier compatibility

### **1\. External Integrations**

* Real SMS provider SDK integration  
* Actual communication with external SMS gateways  
* Real webhook endpoints exposed over network

👉 All external interactions should be **simulated only**

### **2\. API Layer & UI**

* REST / GraphQL API design  
* API Gateway integration  
* Authentication / Authorization  
* Admin UI or management dashboards

👉 Use **CLI or simple triggers** instead

### **3\. Data Persistence**

* No real database setup required  
* No schema design for production storage  
* No migration strategy

👉 Use **in-memory storage only**

### **4\. Infrastructure & Runtime Concerns**

The following must **not influence your design**:

* Message queue systems (Kafka, RabbitMQ, etc.)  
* Background workers / job schedulers  
* Distributed system deployment topology  
* Threading / concurrency model  
* Scaling strategies (horizontal/vertical)  
* Cloud infrastructure (AWS, Azure, GCP)  
* Kubernetes / container orchestration

👉 The focus is **domain design**, not runtime execution

### **5\. Reliability Engineering**

* Retry infrastructure (queue-based retry, DLQ, etc.)  
* Circuit breaker / rate limiting  
* Fault tolerance at infrastructure level  
* Exactly-once / at-least-once delivery guarantees

👉 Only model behavior at **domain level if needed**

### **6\. Performance Optimization**

* High throughput design  
* Latency optimization  
* Load balancing  
* Caching strategies

👉 No performance tuning is required

### **7\. Security & Compliance**

* Encryption  
* PII handling  
* GDPR / compliance concerns  
* Secure credential storage

👉 Not relevant for this challenge

### **8\. Monitoring & Observability Platforms**

* OpenTelemetry / tracing systems  
* Metrics platforms (Prometheus, Datadog, etc.)  
* Log aggregation systems (ELK, etc.)

👉 Basic logging is sufficient

## **⚠️ Important Clarification**

This challenge focuses on:

**Domain design, abstraction, and adaptability**

NOT on:

**Infrastructure, deployment, or runtime execution**

## **📊 Evaluation Criteria**

Your solution will be evaluated based on the following criteria.

The goal is not only to make the system work, but to ensure it remains **clear, maintainable, and adaptable as business requirements evolve**.

### **1\. Core Functionality (15 pts)**

Your system should:

* Accept SMS requests (via CLI or equivalent)  
* Determine carrier (simulation is acceptable)  
* Select a provider based on the given rules  
* Execute the SMS lifecycle correctly

**Expectation:**

* The system behaves correctly for the defined scenarios  
* State transitions are applied properly

### **2\. Code Quality & Structure (10 pts)**

Your implementation should demonstrate:

* Clear and readable code  
* Meaningful naming  
* Proper separation of responsibilities  
* Minimal duplication

**Expectation:**

* The codebase is easy to understand and maintain

### **3\. Provider Integration Design (15 pts)**

Your design should:

* Support multiple SMS providers  
* Clearly separate provider-specific logic from core logic

**Expectation:**

* Adding a new provider should be straightforward  
* Existing logic should not need major modification

### **4\. Routing Behavior (15 pts)**

Your system should:

* Select providers based on country and carrier  
* Apply routing rules correctly

**Expectation:**

* Routing logic should remain understandable and maintainable  
* The system should handle variations in routing rules

### **5\. State Management (15 pts)**

Your system should:

* Represent SMS lifecycle states clearly  
* Support valid transitions  
* Maintain traceability of state changes

**Expectation:**

* The state model remains clear as additional states or transitions are introduced

### **6\. Adaptability to Change (20 pts)**

The system is expected to operate in an evolving business environment.

Your design should be able to handle changes such as:

* New countries  
* New providers  
* Changes in routing rules  
* Additional lifecycle states

**Expectation:**

* Changes can be introduced without significantly restructuring the system  
* The system remains stable and maintainable over time

### **7\. Reusability Across Domains (5 pts)**

The SMS capability should:

* Be usable by multiple domains (e.g., Customer, Booking, ERP)  
* Not be tightly coupled to a single use case

**Expectation:**

* The solution can be reused without domain-specific modifications

### **8\. Observability & Traceability (5 pts)**

Your system should provide visibility into:

* Provider selection  
* State transitions  
* Cost (estimated and actual)

**Expectation:**

* SMS processing can be traced and understood

### **9\. Bonus – Architectural Thinking (+5 pts)**

Additional credit may be given for:

* Clear separation of responsibilities  
* Thoughtful handling of evolving requirements  
* Design choices that support long-term maintainability

## **⚠️ Important Notes**

* Focus on **clarity and adaptability**, not just correctness  
* Avoid unnecessary complexity or over-engineering  
* Infrastructure concerns (queues, scaling, deployment) are **out of scope**

## **🎯 Final Guidance**

A strong solution will:

* Work correctly for current requirements  
* Remain clear and maintainable  
* Continue to function well as requirements evolve

### **💡 Tip for Candidates**

When making design decisions, consider:

“What happens if the requirements change?”
