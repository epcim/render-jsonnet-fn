
local k = import "github.com/jsonnet-libs/k8s-libsonnet/1.24/main.libsonnet";

local utils = import "../vendor/kubernetes-mixin/lib/utils.libsonnet"; 
local conf = (import "../vendor/kubernetes-mixin/mixin.libsonnet") 
+ 
{ 
  _config+:: { 
    kubeStateMetricsSelector: 'job="kube-state-metrics"', 
    cadvisorSelector: 'job="kubernetes-cadvisor"', 
    nodeExporterSelector: 'job="kubernetes-node-exporter"', 
    kubeletSelector: 'job="kubernetes-kubelet"', 
    grafanaK8s+:: { 
      dashboardNamePrefix: 'Mixin / ', 
      dashboardTags: ['kubernetes', 'infrastucture'], 
    }, 
    prometheusAlerts+:: { 
        groups: ['kubernetes-resources'], 
    },
  }, 
} 
+ { 
prometheusAlerts+:: { 
    /*  From Alerts select only the 'kubernetes-system': 
        https://github.com/kubernetes-monitoring/kubernetes-mixin/blob/master/alerts/system_alerts.libsonnet#L9 
    */ 
    groups: 
        std.filter( 
            function(alertGroup) 
              alertGroup.name == 'kubernetes-system' 
            , super.groups 
        ) 
    } 
} 
+ { 
prometheusRules+::  { 
    /*  From Rules select only the 'k8s.rules': 
        https://github.com/kubernetes-monitoring/kubernetes-mixin/blob/master/rules/apps.libsonnet#L10 
    */ 
    groups: 
        std.filter( 
            function(alertGroup) 
              alertGroup.name == 'k8s.rules' 
            , super.groups 
        ) 
    } 
} 
+ { 
    grafanaDashboards+:: {} 
} 
; 

// Manifests as ResourceList
{ resourceList: {
      kind: 'ResourceList',
      items: []
        + [ k.core.v1.configMap.new(name="foo", data=conf.prometheusAlerts) ]
        + [ k.core.v1.configMap.new(name="bee", data=conf.prometheusRules) ]
  }
}


