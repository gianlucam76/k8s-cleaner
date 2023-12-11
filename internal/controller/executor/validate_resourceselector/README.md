If you want to validate your evaluate functions:

1. create a sub-directory
2. create a file named __cleaner.yaml__ containing your Cleaner instance
3. create a file named __matching.yaml__ containing a resource that matches your __Cleaner.ResourcePolicySet.ResourceSelector__
4. create a file named __non-matching.yaml__ containing a resource that does not matches your __Cleaner.ResourcePolicySet.ResourceSelector__
5. run ``make test``

That will run the exact code Cleaner will run in your cluster. 
If you see no error, your Cleaner instance is correct

**This only validates resource __Cleaner.ResourcePolicySet.ResourceSelector__**

If you need to validate your transform function follow instruction in __validate_transform__ directory