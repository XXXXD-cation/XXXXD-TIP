 // 创建唯一性约束
CREATE CONSTRAINT ioc_value_constraint IF NOT EXISTS ON (i:Ioc) ASSERT i.value IS UNIQUE;
CREATE CONSTRAINT threatactor_id_constraint IF NOT EXISTS ON (t:ThreatActor) ASSERT t.id IS UNIQUE;
CREATE CONSTRAINT threatactor_name_constraint IF NOT EXISTS ON (t:ThreatActor) ASSERT t.name IS UNIQUE;
CREATE CONSTRAINT malware_id_constraint IF NOT EXISTS ON (m:Malware) ASSERT m.id IS UNIQUE;
CREATE CONSTRAINT malware_name_constraint IF NOT EXISTS ON (m:Malware) ASSERT m.name IS UNIQUE;
CREATE CONSTRAINT vulnerability_id_constraint IF NOT EXISTS ON (v:Vulnerability) ASSERT v.id IS UNIQUE;
CREATE CONSTRAINT vulnerability_cve_id_constraint IF NOT EXISTS ON (v:Vulnerability) ASSERT v.cveId IS UNIQUE;
CREATE CONSTRAINT ttp_technique_id_constraint IF NOT EXISTS ON (t:Ttp) ASSERT t.techniqueId IS UNIQUE;
CREATE CONSTRAINT campaign_id_constraint IF NOT EXISTS ON (c:Campaign) ASSERT c.id IS UNIQUE;
CREATE CONSTRAINT case_id_constraint IF NOT EXISTS ON (c:Case) ASSERT c.id IS UNIQUE;
CREATE CONSTRAINT tag_id_constraint IF NOT EXISTS ON (t:Tag) ASSERT t.id IS UNIQUE;
CREATE CONSTRAINT tag_name_constraint IF NOT EXISTS ON (t:Tag) ASSERT t.name IS UNIQUE;
CREATE CONSTRAINT source_id_constraint IF NOT EXISTS ON (s:Source) ASSERT s.id IS UNIQUE;
CREATE CONSTRAINT source_name_constraint IF NOT EXISTS ON (s:Source) ASSERT s.name IS UNIQUE;

// 创建索引以提高查询性能
CREATE INDEX ioc_type_index IF NOT EXISTS FOR (i:Ioc) ON (i.type);
CREATE INDEX ioc_status_index IF NOT EXISTS FOR (i:Ioc) ON (i.status);
CREATE INDEX threatactor_aliases_index IF NOT EXISTS FOR (t:ThreatActor) ON (t.aliases);
CREATE INDEX malware_family_index IF NOT EXISTS FOR (m:Malware) ON (m.family);
CREATE INDEX malware_type_index IF NOT EXISTS FOR (m:Malware) ON (m.type);
CREATE INDEX case_title_index IF NOT EXISTS FOR (c:Case) ON (c.title);
CREATE INDEX case_status_index IF NOT EXISTS FOR (c:Case) ON (c.status);

// 添加一些示例节点和关系
// Tag节点
MERGE (t1:Tag {id: "tag1", name: "High"})
MERGE (t2:Tag {id: "tag2", name: "Verified"})
MERGE (t3:Tag {id: "tag3", name: "False Positive"});

// TTP节点
MERGE (ttp1:Ttp {techniqueId: "T1566", name: "Phishing", description: "Phishing is a method of fraudulently obtaining private information..."})
MERGE (ttp2:Ttp {techniqueId: "T1190", name: "Exploit Public-Facing Application", description: "Adversaries may exploit vulnerabilities in public-facing applications..."});

// 注意：实际生产环境中可能不会在初始脚本中添加示例数据，而是通过应用程序或特定的数据导入工具来填充图数据库