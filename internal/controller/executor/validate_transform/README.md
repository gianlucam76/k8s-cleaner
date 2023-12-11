If you want to validate your transform functions:

1. create a sub-directory
2. create a file named __cleaner.yaml__ containing your Cleaner instance
3. create a file named __matching.yaml__ containing a resource that matches your __Cleaner.ResourcePolicySet.ResourceSelector__
4. create a file named __updated.yaml__ containing the expected resource after __Cleaner.ResourcePolicySet.Tranform__ is executed
5. run ``make test``

That will run the exact code Cleaner will run in your cluster. 
If you see no error, your Cleaner instance is correct

**This validates both __Cleaner.ResourcePolicySet.ResourceSelector__ and __Cleaner.ResourcePolicySet.Transform__**

If you need to validate your aggregatedSelection function follow instruction in __validate_aggregatedselection__ directory