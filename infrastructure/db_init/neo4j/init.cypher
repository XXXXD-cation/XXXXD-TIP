// 威胁情报平台(TIP) Neo4j 初始化脚本

// ===== 约束设置 =====
// 设置节点唯一性约束，确保关键实体的唯一识别

// IOC约束 - 确保每种类型的IOC值唯一
CREATE CONSTRAINT ioc_ip_unique IF NOT EXISTS ON (i:IOC:IP) ASSERT i.value IS UNIQUE;
CREATE CONSTRAINT ioc_domain_unique IF NOT EXISTS ON (i:IOC:Domain) ASSERT i.value IS UNIQUE;
CREATE CONSTRAINT ioc_url_unique IF NOT EXISTS ON (i:IOC:URL) ASSERT i.value IS UNIQUE;
CREATE CONSTRAINT ioc_hash_unique IF NOT EXISTS ON (i:IOC:Hash) ASSERT i.value IS UNIQUE;
CREATE CONSTRAINT ioc_email_unique IF NOT EXISTS ON (i:IOC:Email) ASSERT i.value IS UNIQUE;

// 威胁主体约束
CREATE CONSTRAINT threat_actor_unique IF NOT EXISTS ON (t:ThreatActor) ASSERT t.name IS UNIQUE;
CREATE CONSTRAINT malware_unique IF NOT EXISTS ON (m:Malware) ASSERT m.name IS UNIQUE;
CREATE CONSTRAINT campaign_unique IF NOT EXISTS ON (c:Campaign) ASSERT c.name IS UNIQUE;

// 案例约束
CREATE CONSTRAINT case_unique IF NOT EXISTS ON (c:Case) ASSERT c.id IS UNIQUE;

// 事件约束
CREATE CONSTRAINT event_unique IF NOT EXISTS ON (e:Event) ASSERT e.id IS UNIQUE;

// 组织约束
CREATE CONSTRAINT organization_unique IF NOT EXISTS ON (o:Organization) ASSERT o.name IS UNIQUE;

// ===== 索引创建 =====
// 设置索引以提高查询性能

// 时间索引
CREATE INDEX ioc_first_seen IF NOT EXISTS FOR (i:IOC) ON (i.first_seen);
CREATE INDEX ioc_last_seen IF NOT EXISTS FOR (i:IOC) ON (i.last_seen);
CREATE INDEX event_timestamp IF NOT EXISTS FOR (e:Event) ON (e.timestamp);

// 标签索引
CREATE INDEX ioc_tags IF NOT EXISTS FOR (i:IOC) ON (i.tags);
CREATE INDEX threat_actor_tags IF NOT EXISTS FOR (t:ThreatActor) ON (t.tags);
CREATE INDEX malware_tags IF NOT EXISTS FOR (m:Malware) ON (m.tags);

// 属性索引
CREATE INDEX ioc_severity IF NOT EXISTS FOR (i:IOC) ON (i.severity);
CREATE INDEX ioc_confidence IF NOT EXISTS FOR (i:IOC) ON (i.confidence);

// ===== 初始数据 =====
// 创建基础图谱结构和示例数据

// 示例威胁组织
MERGE (apt28:ThreatActor {name: 'APT28', description: '由俄罗斯支持的高级威胁组织', confidence: 'High'})
SET apt28.tags = ['Russia', 'State-sponsored', 'Espionage']
SET apt28.first_seen = '2014-01-01'
SET apt28.last_updated = datetime();

MERGE (apt29:ThreatActor {name: 'APT29', description: '另一个由俄罗斯支持的高级威胁组织', confidence: 'High'})
SET apt29.tags = ['Russia', 'State-sponsored', 'Espionage', 'Data_theft']
SET apt29.first_seen = '2015-03-15'
SET apt29.last_updated = datetime();

// 示例恶意软件
MERGE (sofacy:Malware {name: 'Sofacy', description: 'APT28使用的恶意软件', type: 'Backdoor'})
SET sofacy.tags = ['Backdoor', 'Data_exfiltration', 'Command_and_control']
SET sofacy.first_seen = '2015-06-12'
SET sofacy.last_updated = datetime();

MERGE (cozyBear:Malware {name: 'CozyBear', description: 'APT29使用的恶意软件', type: 'RAT'})
SET cozyBear.tags = ['RAT', 'Persistence', 'Data_theft']
SET cozyBear.first_seen = '2016-02-28'
SET cozyBear.last_updated = datetime();

// 示例活动
MERGE (op2016:Campaign {name: '2016选举干预活动', description: '针对2016年选举的干预活动'})
SET op2016.start_date = '2016-01-01'
SET op2016.end_date = '2016-11-30'
SET op2016.status = 'Completed'
SET op2016.last_updated = datetime();

// 示例IOC
MERGE (ip1:IOC:IP {value: '185.86.148.227', type: 'ipv4'})
SET ip1.first_seen = '2016-03-10'
SET ip1.last_seen = '2016-09-22'
SET ip1.severity = 'High'
SET ip1.confidence = 'High'
SET ip1.tags = ['C2', 'APT28']
SET ip1.last_updated = datetime();

MERGE (domain1:IOC:Domain {value: 'timeservice[.]org', type: 'domain'})
SET domain1.first_seen = '2016-04-05'
SET domain1.last_seen = '2016-10-11'
SET domain1.severity = 'High'
SET domain1.confidence = 'High'
SET domain1.tags = ['C2', 'APT28', 'Sofacy']
SET domain1.last_updated = datetime();

MERGE (hash1:IOC:Hash {value: '0e0294de0802a00e91501e74a7d9bbbed5652f908a2e9636d7e8c8c7222d0a2a', type: 'sha256'})
SET hash1.first_seen = '2016-03-15'
SET hash1.last_seen = '2016-11-20'
SET hash1.severity = 'Critical'
SET hash1.confidence = 'High'
SET hash1.tags = ['Malware', 'Sofacy', 'APT28']
SET hash1.last_updated = datetime();

// 建立实体间关系
// 威胁组织使用恶意软件
MERGE (apt28)-[:USES {first_seen: '2015-06-12'}]->(sofacy)
MERGE (apt29)-[:USES {first_seen: '2016-02-28'}]->(cozyBear)

// 威胁组织发起活动
MERGE (apt28)-[:CONDUCTS {role: 'Primary Actor'}]->(op2016)
MERGE (apt29)-[:CONDUCTS {role: 'Supporting Actor'}]->(op2016)

// IOC关联到恶意软件
MERGE (ip1)-[:ASSOCIATED_WITH {relationship_type: 'C2 Server'}]->(sofacy)
MERGE (domain1)-[:ASSOCIATED_WITH {relationship_type: 'C2 Domain'}]->(sofacy)
MERGE (hash1)-[:IDENTIFIES {relationship_type: 'Binary Hash'}]->(sofacy)

// IOC关联到威胁组织
MERGE (ip1)-[:ATTRIBUTED_TO {confidence: 'High'}]->(apt28)
MERGE (domain1)-[:ATTRIBUTED_TO {confidence: 'High'}]->(apt28)
MERGE (hash1)-[:ATTRIBUTED_TO {confidence: 'High'}]->(apt28)

// 将恶意软件关联到活动
MERGE (sofacy)-[:USED_IN {first_seen: '2016-03-10'}]->(op2016)
MERGE (cozyBear)-[:USED_IN {first_seen: '2016-03-22'}]->(op2016)