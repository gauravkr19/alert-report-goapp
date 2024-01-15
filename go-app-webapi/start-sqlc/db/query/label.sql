	-- name: DisplayTable :many
    SELECT a.id, a.startsat, a.endsat, a.status, ct.alertname, ct.namespace, ct.priority, ct.severity, ct.deployment, ct.pod, ct.container, ct.replicaset FROM alert a 
	LEFT JOIN (
	SELECT * FROM 
	CROSSTAB('select ct.alertid, ct.label, ct.value FROM AlertLabel ct 
	ORDER BY ct.alertid',
	'VALUES (''alertname''), (''namespace''), (''priority''), (''severity''), (''deployment''), (''pod''), (''container''), (''replicaset'')')
	AS ct (alertid int, alertname VARCHAR, namespace VARCHAR, priority VARCHAR, severity VARCHAR, deployment VARCHAR, pod VARCHAR, container VARCHAR, replicaset VARCHAR) 
	) AS ct ON ct.alertid = a.id;