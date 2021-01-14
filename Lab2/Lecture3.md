# Lecture 3: GFS

### Why it is hard to keep global file system?
* **Performance**: Data stored in distributed system are horizontally partitioned on separate server instances, to spread load. This is called sharding.
* **Fault Tolerance**: When large number of machines are involved, there will always be some make mistakes. We need **replications** for restoration.
* **Replication**: Use of replication may result in **inconsistency**, to maintain the **consistency**, the system might result in **low performance**.

### Introduction to GFS:
* Definition to **volatile and non-volatile memory**: Volatile Memory is used to store computer programs and data that CPU needs in real time and is erased once computer is switched off. RAM and Cache memory are volatile memory. Where as Non-volatile memory is static and remains in the computer even if computer is switched off. ROM and HDD are non-volatile memory.
* Basic operations of GFS:
  - Fundamental operations of all file systems, including **create**, **delete**, **open**, **close**, **read**, and **write**.
  - **Snapshot**: creates a copy of a file or a directory tree at low cost. 
  - **Record Append**: allows multiple clients to append data to the same file *concurrently* while guaranteeing the *atomicity* of each individual client’s append.
* GFS Architecture
  | ![Overview of the GFS Architecture](Images/Architecture.png) | 
  |:--:| 
  - A **single master** and **multiple chunk servers**. 
  - **Master** stores directoriies and filenames in tree structrue.
  - **Chunk servers** store all chunk of data using linux file system.
  - Each **file** is divided into *fixed-size* **chunks**.
  - Each **chunk** is identified by an *immutable and globally unique* 64 bit **chunk handle** assigned by the master *at the time of chunk creation*.
  - For reliability, **each chunk** is replicated on 3 other chunkservers by default, though users can designate different replication levels.
* GFS Read Operation
  - Notice that the **clients are not applications**. If applications are trying to read several bytes of data and store those data in a buffer, the application will provide the corresponding filename and byte offsets to the client, and the client will then do the execution.
* GFS Write Operation
  - The majority type of write in GFS is called record appends. In order to keep the atomicity of writing operations, GFS will always try to append some data to all replicas. If some append operations failed, the whole operation will be regarded as a failure and re-executed. As a result, GFS may insert padding or record duplicates in chunks, those areas are considered to be **inconsistent** and are typically **dwarfed** by the amount of user data.
* Major restrictions to GFS system: single master
  - The restricted number of RAMs we can have in a single master makes it difficult to keep information of tremendous number of files. 
  - The execution speed of a single CPU limited the number of requests the master could deal with in one second.
