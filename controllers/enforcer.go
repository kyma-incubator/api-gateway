package controllers

type Enforcer struct {
}


type ObjectFunc func()

func (e *Enforcer) CreateObject(of ObjectFunc) *Enforcer {
}

enforcer.Create(vs).Retries(3).OnFail(vsFailFunc)
enforcer.Update(ar1).Retries(1).OnFailContinue()
enforcer.CreateOrUpdate(ar2).Retries(1).OnFail(arFailFunc)
