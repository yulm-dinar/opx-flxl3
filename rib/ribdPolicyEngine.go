// ribdPolicyEngine.go
package main

import (
     "ribd"
	 "utils/patriciaDB"
	 "l3/rib/ribdCommonDefs"
)
func policyEngineActionRejectRoute(route ribd.Routes, params interface{}) {
    logger.Println("policyEngineActionRejectRoute")
	routeInfo := params.(RouteParams)
	_, err := routeServiceHandler.DeleteV4Route(routeInfo.destNetIp, routeInfo.networkMask, routeInfo.routeType)
	if err != nil {
		logger.Println("deleting v4 route failed with err ", err)
		return
	}
}
func policyEngineActionAcceptRoute(route ribd.Routes, params interface{}) {
    logger.Println("policyEngineActionAcceptRoute")
	routeInfo := params.(RouteParams)
	_, err := createV4Route(routeInfo.destNetIp, routeInfo.networkMask, routeInfo.metric, routeInfo.nextHopIp, routeInfo.nextHopIfType, routeInfo.nextHopIfIndex, routeInfo.routeType, routeInfo.createType, routeInfo.sliceIdx)
	if err != nil {
		logger.Println("creating v4 route failed with err ", err)
		return
	}
}
func policyEngineActionRedistribute(route ribd.Routes, redistributeActionInfo RedistributeActionInfo, params interface {}) {
	logger.Println("policyEngineActionRedistribute")
	//Send a event based on target protocol
    RouteInfo := params.(RouteParams) 
	if ((RouteInfo.createType != Invalid || RouteInfo.deleteType != Invalid ) && redistributeActionInfo.redistribute == false) {
		logger.Println("Don't redistribute action set for a route create/delete, return")
		return
	}
	var evt int
	if RouteInfo.createType != Invalid {
		logger.Println("Create type not invalid")
		evt = ribdCommonDefs.NOTIFY_ROUTE_CREATED
	} else if RouteInfo.deleteType != Invalid {
		logger.Println("Delete type not invalid")
		evt = ribdCommonDefs.NOTIFY_ROUTE_DELETED
	} else {
		logger.Println("Create/Delete invalid, redistributeAction set to ", redistributeActionInfo.redistribute)
		if redistributeActionInfo.redistribute == true {
			logger.Println("evt = NOTIFY_ROUTE_CREATED")
			evt = ribdCommonDefs.NOTIFY_ROUTE_CREATED
		} else
		{
			logger.Println("evt = NOTIFY_ROUTE_DELETED")
			evt = ribdCommonDefs.NOTIFY_ROUTE_DELETED
		}
	}
    switch redistributeActionInfo.redistributeTargetProtocol {
      case ribdCommonDefs.BGP:
        logger.Println("Redistribute to BGP")
        RouteNotificationSend(RIBD_BGPD_PUB, route, evt)
        break
      default:
        logger.Println("Unknown target protocol")	
    }
}
func policyEngineActionUndoRedistribute(route ribd.Routes, redistributeActionInfo RedistributeActionInfo, params interface {}) {
	logger.Println("policyEngineActionUndoRedistribute")
	//Send a event based on target protocol
	var evt int
	logger.Println("redistributeAction set to ", redistributeActionInfo.redistribute)
	if redistributeActionInfo.redistribute == true {
	   logger.Println("evt = NOTIFY_ROUTE_DELETED")
	   evt = ribdCommonDefs.NOTIFY_ROUTE_DELETED
	} else {
		logger.Println("evt = NOTIFY_ROUTE_CREATED")
		evt = ribdCommonDefs.NOTIFY_ROUTE_CREATED
	}
    switch redistributeActionInfo.redistributeTargetProtocol {
      case ribdCommonDefs.BGP:
        logger.Println("Redistribute to BGP")
        RouteNotificationSend(RIBD_BGPD_PUB, route, evt)
        break
      default:
        logger.Println("Unknown target protocol")	
    }
}

func policyEngineUndoActions(route ribd.Routes, policyStmt PolicyStmt, params interface{}) {
	logger.Println("policyEngineUndoActions")
	if policyStmt.actions == nil {
		logger.Println("No actions")
		return
	}
	var i int
	for i=0;i<len(policyStmt.actions);i++ {
	  logger.Printf("Find policy action number %d name %s in the action database\n", i, policyStmt.actions[i])
	  actionItem := PolicyActionsDB.Get(patriciaDB.Prefix(policyStmt.actions[i]))
	  if actionItem == nil {
	     logger.Println("Did not find action ", policyStmt.actions[i], " in the action database")	
		 continue
	  }
	  action := actionItem.(PolicyAction)
	  logger.Printf("policy action number %d type %d\n", i, action.actionType)
		switch action.actionType {
		   case ribdCommonDefs.PolicyActionTypeRouteDisposition:
		      logger.Println("PolicyActionTypeRouteDisposition action to be applied")
			  logger.Println("RouteDisposition action = ", action.actionInfo)
			  if action.actionInfo.(string) == "Accept" {
                 logger.Println("Accept action - undoing it by deleting")
				 policyEngineActionRejectRoute(route, params)
				 return
			  }
			  break
		   case ribdCommonDefs.PolicyActionTypeRouteRedistribute:
		      logger.Println("PolicyActionTypeRouteRedistribute action to be applied")
			  policyEngineActionUndoRedistribute(route, action.actionInfo.(RedistributeActionInfo), params)
			  break
		   default:
		      logger.Println("Unknown type of action")
			  return
		}
	}
}
func policyEngineImplementActions(route ribd.Routes, policyStmt PolicyStmt, params interface {}) {
	logger.Println("policyEngineImplementActions")
	if policyStmt.actions == nil {
		logger.Println("No actions")
		return
	}
	var i int
	createRoute := false
	for i=0;i<len(policyStmt.actions);i++ {
	  logger.Printf("Find policy action number %d name %s in the action database\n", i, policyStmt.actions[i])
	  actionItem := PolicyActionsDB.Get(patriciaDB.Prefix(policyStmt.actions[i]))
	  if actionItem == nil {
	     logger.Println("Did not find action ", policyStmt.actions[i], " in the action database")	
		 continue
	  }
	  action := actionItem.(PolicyAction)
	  logger.Printf("policy action number %d type %d\n", i, action.actionType)
		switch action.actionType {
		   case ribdCommonDefs.PolicyActionTypeRouteDisposition:
		      logger.Println("PolicyActionTypeRouteDisposition action to be applied")
			  logger.Println("RouteDisposition action = ", action.actionInfo)
			  if action.actionInfo.(string) == "Reject" {
                 logger.Println("Reject action")
				 policyEngineActionRejectRoute(route, params)
				 return
			  }
			  createRoute = true
			  break
		   case ribdCommonDefs.PolicyActionTypeRouteRedistribute:
		      logger.Println("PolicyActionTypeRouteRedistribute action to be applied")
			  policyEngineActionRedistribute(route, action.actionInfo.(RedistributeActionInfo), params)
			  break
		   default:
		      logger.Println("Unknown type of action")
			  return
		}
	}
	logger.Println("createRoute = ",createRoute)
	if createRoute {
		policyEngineActionAcceptRoute(route, params)
	}
}
func PolicyEngineMatchConditions(route ribd.Routes, policyStmt PolicyStmt) (match bool){
    logger.Println("policyEngineMatchConditions")
	var i int
	allConditionsMatch := true
	anyConditionsMatch := false
	for i=0;i<len(policyStmt.conditions);i++ {
	  logger.Printf("Find policy condition number %d name %s in the condition database\n", i, policyStmt.conditions[i])
	  conditionItem := PolicyConditionsDB.Get(patriciaDB.Prefix(policyStmt.conditions[i]))
	  if conditionItem == nil {
	     logger.Println("Did not find condition ", policyStmt.conditions[i], " in the condition database")	
		 continue
	  }
	  condition := conditionItem.(PolicyCondition)
	  logger.Printf("policy condition number %d type %d\n", i, condition.conditionType)
      switch condition.conditionType {
		case ribdCommonDefs.PolicyConditionTypePrefixMatch:
		  logger.Println("PolicyConditionTypePrefixMatch case")
		break
		case ribdCommonDefs.PolicyConditionTypeProtocolMatch:
		  logger.Println("PolicyConditionTypeProtocolMatch case")
		  matchProto := condition.conditionInfo.(int)
		  if matchProto == int(route.Prototype) {
			logger.Println("Protocol condition matches")
			anyConditionsMatch = true
		  } else {
			logger.Println("Protocol condition does not match")
			allConditionsMatch = false
		  } 
		break
		default:
		  logger.Println("Not a known condition type")
          return match
	  }
	}
   if policyStmt.matchConditions == "all" && allConditionsMatch == true {
	return true
   }
   if policyStmt.matchConditions == "any" && anyConditionsMatch == true {
	return true
   }
   return match
}

func policyEngineApplyPolicy(route *ribd.Routes, policyStmt PolicyStmt, params interface{}) {
	logger.Println("PolicyEngineApplyPolicy - ", policyStmt.name)
	if policyStmt.conditions == nil {
		logger.Println("No policy conditions")
		return
	}
	match := PolicyEngineMatchConditions(*route, policyStmt)
	logger.Println("match = ", match)
	if !match {
		logger.Println("Conditions do not match")
		return
	}
	policyEngineImplementActions(*route, policyStmt, params)
    routeInfo := params.(RouteParams)
	var op int
	if routeInfo.deleteType != Invalid {
		op = del
	} else {
		op = add
	    route.PolicyHitCounter++
	    updateRoutePolicyState(*route, op, policyStmt.name)
	}
	updatePolicyRouteMap(*route, policyStmt, op)
}
func policyEngineCheckPolicy(route *ribd.Routes, params interface {}) {
	logger.Println("policyEngineCheckPolicy")
	
	//Protocol based policy checks
	policyList := ProtocolPolicyListDB[int(route.Prototype)]
	if(policyList == nil) {
		logger.Println("No policy configured for this route type ", route.Prototype)
		//return 0, err
	}
	logger.Printf("Number of policies configured for this route type %d is %d\n", route.Prototype, len(policyList))
	for policyNum :=0;policyNum < len(policyList);policyNum++ {
		logger.Printf("policy number %d name %s\n", policyNum, policyList[policyNum])
//		policyEngineApplyPolicy(route, policyList[policyNum], params)
	}
	
	//Prefix based policy checks
}
func PolicyEngineFilter(route ribd.Routes, policyPath int, params interface{}) {
	logger.Println("PolicyEngineFilter")
	var policyPath_Str string
	idx :=0
	var policyInfo interface{}
//	policyEngineCheckPolicy(route, policyPath, funcName, params)
    routeInfo := params.(RouteParams)
	logger.Println("Beginning createType = ", routeInfo.createType, " deleteType = ", routeInfo.deleteType)
	for ;; {
       if route.PolicyList != nil {
		  if idx >= len(route.PolicyList) {
			break
		  }
		  logger.Println("getting policy stmt ", idx, " from route.PolicyList")
	      policyInfo = 	PolicyDB.Get(patriciaDB.Prefix(route.PolicyList[idx]))
		  idx++
	   } else if routeInfo.deleteType != Invalid {
		  logger.Println("route.PolicyList empty and this is a delete operation for the route, so break")
          break
	   } else if localPolicyStmtDB == nil {
		  logger.Println("localPolicyStmt nil")
			//case when no policies have been applied to the route
			//need to apply the default policy
		   break	   
		} else {
            if idx >= len(localPolicyStmtDB) {
				break
			}		
		    logger.Println("getting policy stmt ", idx, " from localPolicyStmtDB")
            policyInfo = PolicyDB.Get(localPolicyStmtDB[idx].prefix)
			idx++
	   }
	   if policyInfo == nil {
	      logger.Println("Nil policy")
		  continue
	   }
	   policyStmt := policyInfo.(PolicyStmt)
	   if policyPath == ribdCommonDefs.PolicyPath_Import {
	      policyPath_Str = "Import"
	   } else {
	        policyPath_Str = "Export"
	   }
	   if policyPath == ribdCommonDefs.PolicyPath_Import && policyStmt.importPolicy == false || 
	      policyPath == ribdCommonDefs.PolicyPath_Export && policyStmt.exportPolicy == false {
	         logger.Println("Cannot apply the policy ", policyStmt.name, " as ", policyPath_Str, " policy")
			 continue
	   }
	   policyEngineApplyPolicy(&route, policyStmt, params)
	}
/*	if localPolicyStmtDB == nil {
		logger.Println("No policies configured, so accept the route")
        //should be replaced by default import policy action
	} else {
		for idx :=0;idx < len(localPolicyStmtDB);idx++ {
		//for idx :=0;idx < len(policList);idx++ {
			if localPolicyStmtDB[idx].isValid == false {
				continue
			}
			policyInfo := PolicyDB.Get(localPolicyStmtDB[idx].prefix)
			if policyInfo == nil {
				logger.Println("Nil policy")
				continue
			}
			policyStmt := policyInfo.(PolicyStmt)
			if policyPath == ribdCommonDefs.PolicyPath_Import {
				policyPath_Str = "Import"
			} else {
				policyPath_Str = "Export"
			}
			if policyPath == ribdCommonDefs.PolicyPath_Import && policyStmt.importPolicy == false || 
			   policyPath == ribdCommonDefs.PolicyPath_Export && policyStmt.exportPolicy == false {
				logger.Println("Cannot apply the policy ", policyStmt.name, " as ", policyPath_Str, " policy")
				continue
			}
		    policyEngineApplyPolicy(&route, policyStmt, params)
        }
	}*/
	logger.Println("After policyEngineApply policyCounter = ", route.PolicyHitCounter)
	if route.PolicyHitCounter == 0{
		logger.Println("Need to apply default policy, policyPath = ", policyPath, "policyPath_Str= ", policyPath_Str)
		if policyPath == ribdCommonDefs.PolicyPath_Import {
		   logger.Println("Applying default import policy")
		    //TO-DO: Need to add the default policy to policyList of the route
           policyEngineActionAcceptRoute(route , params ) 
		} else if policyPath == ribdCommonDefs.PolicyPath_Export {
			logger.Println("Applying default export policy")
		}
	}
	var op int
	if routeInfo.deleteType != Invalid {
		op = delAll		//wipe out the policyList
	    updateRoutePolicyState(route, op, "")
	} 
}

func policyEngineApplyForRoute(prefix patriciaDB.Prefix, item patriciaDB.Item, handle patriciaDB.Item) (err error) {
   logger.Println("policyEngineApplyForRoute %v", prefix)	
   policy := handle.(PolicyStmt)
   rmapInfoRecordList := item.(RouteInfoRecordList)
   if len(rmapInfoRecordList.routeInfoList) == 0 {
      logger.Println("len(rmapInfoRecordList.routeInfoList) == 0")
	  return err	
   }
   logger.Println("Selected route index = ", rmapInfoRecordList.selectedRouteIdx)
   selectedRouteInfoRecord := rmapInfoRecordList.routeInfoList[rmapInfoRecordList.selectedRouteIdx]
   policyRoute := ribd.Routes{Ipaddr: selectedRouteInfoRecord.destNetIp.String(), Mask: selectedRouteInfoRecord.networkMask.String(), NextHopIp: selectedRouteInfoRecord.nextHopIp.String(), NextHopIfType: ribd.Int(selectedRouteInfoRecord.nextHopIfType), IfIndex: selectedRouteInfoRecord.nextHopIfIndex, Metric: selectedRouteInfoRecord.metric, Prototype: ribd.Int(selectedRouteInfoRecord.protocol)}
   params := RouteParams{destNetIp:policyRoute.Ipaddr, networkMask:policyRoute.Mask, routeType:policyRoute.Prototype, sliceIdx:policyRoute.SliceIdx, createType:Invalid, deleteType:Invalid}
   policyEngineApplyPolicy(&policyRoute, policy, params)
   return err
}
func PolicyEngineTraverseAndApply(policy PolicyStmt) {
	logger.Println("PolicyEngineTraverseAndApply - traverse routing table and apply policy ", policy.name)
    RouteInfoMap.VisitAndUpdate(policyEngineApplyForRoute, policy)
}
func PolicyEngineTraverseAndReverse(policy PolicyStmt) {
	logger.Println("PolicyEngineTraverseAndReverse - traverse routing table and inverse policy actions", policy.name)
	if policy.routeList == nil {
		logger.Println("No route affected by this policy, so nothing to do")
		return
	}
	for idx:=0;idx<len(policy.routeList);idx++ {
		routeInfoRecordListItem := RouteInfoMap.Get(patriciaDB.Prefix(policy.routeList[idx]))
		if routeInfoRecordListItem == nil {
			logger.Println("routeInfoRecordListItem nil for prefix ", policy.routeList[idx])
			continue
		}
		routeInfoRecordList := routeInfoRecordListItem.(RouteInfoRecordList)
        selectedRouteInfoRecord := routeInfoRecordList.routeInfoList[routeInfoRecordList.selectedRouteIdx]
        policyRoute := ribd.Routes{Ipaddr: selectedRouteInfoRecord.destNetIp.String(), Mask: selectedRouteInfoRecord.networkMask.String(), NextHopIp: selectedRouteInfoRecord.nextHopIp.String(), NextHopIfType: ribd.Int(selectedRouteInfoRecord.nextHopIfType), IfIndex: selectedRouteInfoRecord.nextHopIfIndex, Metric: selectedRouteInfoRecord.metric, Prototype: ribd.Int(selectedRouteInfoRecord.protocol)}
        params := RouteParams{destNetIp:policyRoute.Ipaddr, networkMask:policyRoute.Mask, routeType:policyRoute.Prototype, sliceIdx:policyRoute.SliceIdx, createType:Invalid, deleteType:Invalid}
		policyEngineUndoActions(policyRoute, policy, params)
		deleteRoutePolicyState(patriciaDB.Prefix(policy.routeList[idx]), policy.name)
	}
}
