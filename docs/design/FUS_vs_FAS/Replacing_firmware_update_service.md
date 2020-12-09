# <a name="top">Replacing Firmware Update Service</a>
*Andrew Nieuwsma | March 25th, 2020*

In this document I will outline the architectural and engineering weaknesses present in the Firmware Update Service (FUS). I will discuss what steps should be taken to correct these issues and describe what effort of work it will take to complete. I will also examine some of the 'system' at play that brought us to this point, and what actions we have taken to adjust our path.

### <a name="preface">Preface</a>

In February HMS completed work on the Task Runner Service (TRS) http library — a highly parallel http wrapper library. As part of completing that effort we selected FUS as a good candidate to convert its http usage to utilize TRS (to aide in our scalability effort). As part of that implementation effort Sean and I started to get a more detailed exposure to FUS and what it does and identified some deficiencies in the organization of FUS.

As part of my workload I was tasked with leading the HMS resiliency effort. With my recent exposure to FUS, that initial experience led to a more directed inspection of the architecture and structure of the code, which revealed the systemic nature of the disorganization. In response we set out to do refactoring to make the code maintainable and testable, but as our work progressed the scope of change necessary crossed from refactoring (internal reorganization without a contractual change) to re-design/implement of the API and the internals of the service.

These changes necessitate that whatever follows FUSv1 (FUSv2 or FAS(Firmware Action Service)) will NOT be compliant with FUSv1 (we are currently tracking this work in CASM-1830).  This change in interface also means that a more concerted effort must take place to engage our stakeholders to limit the impact of the change in contract (see [Stakeholders](#stakeholders) - for more details). 

## Embracing Agility & Risk
The deficiencies present in FUS have compounded over the life of FUS as the architecture & requirements have evolved.  I want to make it abundantly clear that I don’t blame <strong>ANYONE</strong> for this, because at the time we were kind of in a ‘wild-west’ of building services; and now we are at a place of maturity. There are a number of factors that led to FUS being what it is today, some good, some not so good.  

I think it is essential to lay the ground work for this highly detailed report by providing context and light to help us not to judge, but to understand how we got here. As engineers we take pride in our work, and while we do not seek failure or mistake, we should not, nor cannot be so afraid of failure that we do not take measured risk.  Recently as a CASM organization we were admonished by our leaders in the value of 'failing fast'; this is an opportunity to embrace that challenge and respond strongly.  FUS as it stands today is not a failure, it is the precursor to the next and better version that will be able to go further, and do so better because of what we have learned.

Over the last year our entire organization has been through constant 'white-water' change.  This change has definitely impacted and redirected how we employ our craft, how we work together, and even the nature of the problems we solve. At a high level, and without going into too much detail I believe (after discussion and reflection with the team) the following factors of the 'system' were contributing factors.  

 * Siloing -> FUS was a siloed development project; with our resource constraints and timelines it was natural to split services into small silos. We have embraced the shared ownership model and have a dedicated working group that steward the development of FUS. As we integrated TRS into FUS we converted FUS to being the shared ownership model.  This has allowed us to already make improvements in existing FUS (However the best path, as I will discuss later, is still a new version).
 * Process -> The agile development paradigm provides many challenges and adaptations.  Where waterfall fell short was that 'changes invariably happen' and agile provides a framework for better adaptation. Agile was a new process to us as a team; and while we do not use all the ceremonies of scrum we have embraced the mantra of agility. A 'corruption' of agile is to accrue technical debt because of pressure to release.  Agile is a push model (where mature features are PUSHED towards production).  The previous release model was pull driven, which can cause teams to release software that is not at high enough quality because the integration points are too rigid. 
 * Skills -> Most of us on the team, a year ago, were very new to Go-lang, RESTful development, in a kubernetes environment.  In the past year we have learned many best implementations and anti-patterns and increased our skills and competencies. 
 * Domain leaders -> The architecture team has gone through a lot of change due to corporate reorganization and employee turnover; this led to some resource gaps, where we could validate requirements, objectives, and approaches.

I am happy to say that I work on a learning and evolving team, so while we have not attained 'perfection' we have grown tremendously over the past year. Our whole organization has adapted to this constant change and we are much stronger. 

With this foundation of understanding laid I will constrain myself to focus on the current state of FUS and what our path should be moving forward.  The duration of this report describes the deficiencies of FUS, and what needs to be done to correct them.  I do not discuss the successes of FUS, or the 'positive' things of FUS, but am taking a rather 'clinical' view regarding the state of the architecture and code.

## Table of Contents
  * [Use Cases](#useCases)
  * [Stakeholders](#stakeholders)
  * [Architecture](#architecture)
      * [Rest API](#restApi)
      * [Control Loop](#controlLoop) 
  * [Engineering](#engineering)
      * [Interface Segregation](#interfaceSegregation)
      * [Testability](#testability)
      * [Error Handling](#errorHandling)
      * [Readability & Code-line organization](#readability)
      * [Hand-crafted JSON](#json)
      * [Data Races & Globals](#globals)
      * [Logging](#logging)
      * [Parallelism](#parallelism)
  * [Solutions](#solutions)
      * [Firmware Action Service](#fas)
      * [Current Status of Efforts](#currentStatus) 
      * [Estimated Remaining Work](#estimation)
  * [Conclusion](#conclusion)

## <a name="useCases"> Use Cases </a>

FUS exists to 'update' the firmware of Redfish devices that have been discovered & inventoried as part of the Shasta system.  The main use cases of FUS are:

 1. Perform a firmware action (upgrade, leveling, downgrade), on a set of Redfish Devices as constrained by dependency rules.
 2. Record (and Restore) a 'snapshot' of the firmware states of the system. 
 3. Retrieve current firmware version information from Redfish devices.
 4. Define and enforce device dependency rules necessary to conduct a sequence  of firmware actions. 

A limitation of FUSv1 is that there are only two 'update' semantics for firmware: restore, and latest.  There is not a concept for targeting a specific version, leveling or downgrading.   


## <a name="stakeholders">Stakeholders </a>
A non-exhaustive list of stakeholders:

*This data has been removed due to operational security guidelines*

## <a name="architecture"> Architecture </a>

FUS has two predominate functions: provide a RESTful interface for firmware actions & control/schedule logic that performs the requested actions. 

### <a name="restApi"> REST API </a>
In reviewing the API of FUS there are some limitations, some minor, some more significant. They include:

 * inability for an admin to request a firmware action be halted.
 * inability of the system to stop a 'hung' update.
 * requesting to see if a snapshot exists results in a snapshot being taken.  In this case a GET is not idempotent and is creational (doing what a POST should do).
 * resource pluralization.  All FUS resources are singleton (an anti-pattern).
 * an `/update` results in an `updateID` which can be checked by querying `/status`.  The `/status` resource should not exist, it is actually the `GET` verb of an `/update(s)`.
 * Several of the resources tack on an `.../all` to get everything.  By pluralizing we can leave `/all` off and return everything. 
 * dynamically changing the value of timestamps to contain error messages (instead of using an error object). 
 * the dependencies data model and rest implementation is not functionally correct.  It relies on a partial composite key lookup which can result in misses and is very hard to maintain. Dependencies needs a near complete overhaul to reflect the desired data model.  This in turn will have cascading impacts across the control loop and rest of the software. 

### <a name="controlLoop"> Control Loop </a>
The control loop, for the most part, is stable.  The changes needed to dependencies will have a large effect on the control loop as the data model and data relationships will have to change to reflect the natural structure of the data.   Additionally interface segregation will help contain core domain logic that should not be present in the control loop.  Ex: the control loop gets ALL updates from ALL time that it knows about.  It then sorts through them and determines which one to act upon.  This type of logic still needs to happen but should happen at a lower level so FUS can focus on the core competency of performing an update. 

One of the inherent difficulties of performing firmware actions, or any redfish action is that vendors can implement differently.  This has led to (by necessity) having to know different processes for 'updating' firmware on different devices.  This is still very much necessary, but the actions should be put into a common interface, and type specific implementation should be created.  This would aid to the readability of FUS and by extension is maintainability.

### <a name="complexity"> Domain Complexities</a>
To aid in understanding it is important to have awareness of some of the core domain complexities that FUS has to resolve. These include:

 * Managing the differences between vendor implementations of Redfish.  When FUS was first designed it was believed that each vendor of devices would implement redfish the same way; this assumption has been proven false. 
 * Resolving the 'state chart' and dependencies for how an update sequence is executed. A naive approach to firmware upgrading would not consider other targets that must attain a firmware level in order to allow targets on the same host to upgrade.  Ex) Not being able to upgrade nodeBMC firmware without first upgrading the BIOS firmware. 
 * Managing and expressing firmware versions and hierarchies. As each vendor is responsible to implement Redfish, each vendor also creates a naming scheme for its firmware.  In many cases there is no logical ordering to this firmware, so it is not programmatically possible to divine which firmware is considered 'latest'.  Related to this, vendors can release several firmwares with the same (or mostly same name).  For example they might release `fw1.2.3-prod` & `fw1.2.3-debug` where these both exist at the same 'version' but have different options enabled. 

## <a name="engineering"> Engineering </a>

### <a name="interfaceSegregation"> Interface segregation </a>
FUS has no internal interface segregation.  All of the code is in the same `main` package. While this may seem convenient it has led to cyclic dependencies in the code.  Massive functions of 500+ lines of code exists that handle digesting the API, determining what to do, calling out to storage mechanisms (like etcd).  The lack of boundaries makes it very hard to maintain or test the code.  For example a recent enhancement to use the Task Runner Service (TRS), which should have taken a few hours took close to a full week's worth of developer effort.

FUS does not currently have layers but natural fault lines exists to split apart the functionality of the service. The following boundaries need to be defined in the code:

 * Hardware State Manager interface
 * Storage of Domain Objects interface : dependencies, updates, snapshots
 * API (json, data validation, general HTTP)
 * Domain (aka business logic)
 * Models (aka structs + getters/setters + equality, etc)

### <a name="testability"> Testability </a>
FUS has minimal testing capability.  Unit tests are very difficult/impossible as too much of the code is 'main code' and not enough of it is segregated behind interfaces.  The number of unit tests is around ~20.   As FUS becomes more segregated unit tests will be much more accessible and useful.  Many of the unit tests that do exist are testing things that are outside the application boundary (like testing that etcd can store different characters (that should be tested as part of our HMS etcd library)).

Higher order testing (integration testing, functional testing, & regression testing), does exist to some degree, but it is unclear to what level.  I have not had a chance to focus on those aspects of FUS. 

### <a name="errorHandling"> Error Handling </a>
Go has a built in `errors` type.  FUS does not use it, instead using mostly string values (an anti-pattern). The lack of typed errors makes checking for errors from functions difficult and error prone, and leads to a lot of duplication. Furthermore errors are often ignored, which begs the question of why have them in the first place.

### <a name="readability"> Readability & Code line organization </a>
FUS has very poor code-line organization. It is very difficult to navigate the code base as everything is 'jammed' together.  Additionally there is a deficiency of meaningful names for variables and functions; and a fair amount of 'name reuse' which is confusing.  The tightly coupled nature of the code means that a small change cascades into a large change from the sheer number of files, function, and lines of code that must be touched to propagate the change. 

### <a name="json"> Hand-crafted JSON </a>
Go has a very powerful built in ability to generate JSON.  Some of the structures in FUS were hand crafted, by writing "'s and /'s and manual spaces.  This should never be done in Go as it's just not necessary and leads to non-conformance, troubleshooting and maintenance issues.

### <a name="globals"> Data races & globals </a>
FUS has a number of data races due to the use of a (now deprecated) hms internal http library.  The instances of this library need to be removed from FUS in favor of a standard http interface. Furthermore the way we are using our ETCD library and 'global state' has lead to data races that could lead to silent data corruption. FUS makes extensive use of GLOBALS some of which are non-constant (which is usually considered an anti-pattern).

### <a name="logging"> Logging </a>
FUS relies on the built in go logger, which is fine in and of itself.  I have personally found it to be deficient as you cannot easily do leveled logging (printing `TRACE` level messages vs `DEBUG` level messages).  Go log should be replaced with Logrus (which from my perspective) is the defacto standard as it allows much more fine grained control and consistency. 

### <a name="parallelism"> Parallelism </a>
FUS has a confusing concurrency model.  The complexity of which limits its ability to be extended.  With the creation of TRS, a reasonable amount of FUS actions can be refactored to take advantage of this highly parallel library.  

## <a name="solutions"> Solutions </a>

### <a name="fas">Firmware Action Service </a>
I have previously advocated for a version 2 of FUS.  I think that is insufficient and will un-necessarily constrain us to too many of FUS's short comings.  We should instead use a heavily modified and re-tooled FUS and release it as the `Firmware Action Service`.  A v2 also implies that there is an upgrade path from v1 to v2.  As the APIs are not compliant (nor are their data stores). We would already have to do a wholesale deprecation of FUSv1 in the v1.3 timeline.  

By creating a new service we can still give our consumers (SAT, etc) time to convert to the new version in the v1.3 timeline, and can have a much better product released to our customers. Furthermore this would simplify the 'HELM' aspect of the deployment, as this would be a wholesale replacement, without having to try to make an incremental conversion.  This does eliminate a 'transformation path' from v1 to FAS, but this will save a large amount of effort. 

## <a name="conclusion">Conclusion </a>

Originally we were hoping that we might be able to incrementally fix FUS as needed to correct the problems; but the complexity of FUS & its code organization as it currently stands is too great to accomplish this with a piecemeal approach. Continuing down the path of incremental development would continue compounding the level of disorganization and would make feature implementation and bug fixes harder and harder to achieve.   

We are actively reshaping FUS into a more readable, maintainable, stable service. I am confident that through this concerted effort we can provide a stronger implementation to perform firmware actions, while creating a more resilient platform that can be extended and maintained.